package raft

import (
	"MIT-6_824-2021/labgob"
	"bytes"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"MIT-6_824-2021/labrpc"
)

const (
	Follower  = 0
	Candidate = 1
	Leader    = 2

	ApplyInterval       = time.Millisecond * 5
	HBInterval          = time.Millisecond * 15
	ElectionBaseTimeout = time.Millisecond * 100
)

func ElectionTimeout() time.Duration {
	return ElectionBaseTimeout + time.Duration(rand.Int63())%ElectionBaseTimeout
}

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type Raft struct {
	mu        sync.Mutex
	peers     []*labrpc.ClientEnd
	persister *Persister
	me        int
	dead      int32

	applyCh chan ApplyMsg
	state   int

	currentTerm int
	votedFor    int
	logs        []LogEntry
	grantedVote int
	newLog      int

	commitIndex int
	lastApplied int
	nextIndex   []int
	matchIndex  []int

	lastIncludedIndex int
	lastIncludedTerm  int

	electionTimer *time.Timer
}

type LogEntry struct {
	Term    int
	Command interface{}
}

type RequestVoteArgs struct {
	LastLogTerm  int
	LastLogIndex int
	Term         int
	CandidateId  int
}

type RequestVoteReply struct {
	VoteGranted bool
	Term        int
}

type AppendEntriesArgs struct {
	Term            int
	LeaderId        int
	LastLogIndex    int
	LastLogTerm     int
	Entries         []LogEntry
	LeaderCommitIdx int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
	UpToIdx int
}

type InstallSnapshotArgs struct {
	Term              int
	LeaderId          int
	LastIncludedIndex int
	LastIncludedTerm  int
	Data              []byte
}

type InstallSnapshotReply struct {
	Term int
}

func (rf *Raft) ticker() {
	for rf.killed() == false {
		select {
		case <-rf.electionTimer.C:
			rf.mu.Lock()
			rf.state = Candidate
			rf.currentTerm++
			rf.grantedVote = 1
			rf.votedFor = rf.me
			rf.persist()
			rf.StartElection()
			rf.electionTimer.Reset(ElectionTimeout())
			rf.mu.Unlock()
		}
	}
}

func (rf *Raft) StartElection() {
	for i := range rf.peers {
		if i != rf.me {
			go func(i int) {
				rf.mu.Lock()
				args := RequestVoteArgs{
					LastLogTerm:  rf.getLastTerm(),
					LastLogIndex: rf.getLastIndex(),
					Term:         rf.currentTerm,
					CandidateId:  rf.me,
				}
				reply := RequestVoteReply{}
				rf.mu.Unlock()
				ok := rf.sendRequestVote(i, &args, &reply)
				if ok {
					rf.mu.Lock()
					defer rf.mu.Unlock()
					if rf.state != Candidate || rf.currentTerm > args.Term {
						return
					}
					if reply.Term > args.Term {
						rf.currentTerm = Max(rf.currentTerm, reply.Term)
						rf.state = Follower
						rf.grantedVote = 0
						rf.votedFor = -1
						rf.persist()
					}
					if reply.VoteGranted && rf.currentTerm == args.Term {
						rf.grantedVote++
						if rf.grantedVote > len(rf.peers)/2 {
							rf.state = Leader
							rf.votedFor = -1
							rf.grantedVote = 0
							rf.persist()
							rf.matchIndex = make([]int, len(rf.peers))
							rf.matchIndex[rf.me] = rf.getLastIndex()
							rf.nextIndex = make([]int, len(rf.peers))
							for i := 0; i < len(rf.peers); i++ {
								rf.nextIndex[i] = rf.getLastIndex() + 1
							}
							rf.electionTimer.Reset(ElectionTimeout())
						}
					}
				}
			}(i)
		}
	}
}

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term < rf.currentTerm {
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		return
	}
	reply.Term = rf.currentTerm
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.state = Follower
		rf.grantedVote = 0
		rf.persist()
	}
	if !rf.ValidLog(args.LastLogTerm, args.LastLogIndex) || rf.votedFor != -1 && rf.votedFor != args.CandidateId && reply.Term == args.Term {
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		return
	}
	rf.votedFor = args.CandidateId
	rf.currentTerm = args.Term
	rf.electionTimer.Reset(ElectionTimeout())
	reply.VoteGranted = true
	reply.Term = rf.currentTerm
	rf.persist()
}

func (rf *Raft) applier() {
	retry := 0
	for rf.killed() == false {
		rf.mu.Lock()
		if rf.newLog == 0 && retry < 4 {
			rf.mu.Unlock()
			time.Sleep(HBInterval)
			retry++
			continue
		}
		retry = 0
		rf.newLog = 0
		if rf.state == Leader {
			rf.mu.Unlock()
			rf.leaderApplier()
		} else {
			rf.mu.Unlock()
		}
	}
}

func (rf *Raft) leaderApplier() {
	for i := 0; i < len(rf.peers); i++ {
		if i == rf.me {
			rf.electionTimer.Reset(ElectionTimeout())
			continue
		}
		go func(i int) {
			rf.mu.Lock()
			if rf.state != Leader {
				rf.mu.Unlock()
				return
			}
			if rf.lastIncludedIndex > rf.nextIndex[i]-1 {
				go rf.leaderSendSnapshot(i)
				rf.mu.Unlock()
				return
			}
			lastLogIndex, lastLogTerm := rf.getLastLogInfo(i)
			args := AppendEntriesArgs{
				Term:            rf.currentTerm,
				LeaderId:        rf.me,
				LastLogIndex:    lastLogIndex,
				LastLogTerm:     lastLogTerm,
				LeaderCommitIdx: rf.commitIndex,
			}
			if rf.nextIndex[i] <= rf.getLastIndex() {
				entries := make([]LogEntry, 0)
				entries = append(entries, rf.logs[rf.nextIndex[i]-rf.lastIncludedIndex:]...)
				args.Entries = entries
			} else {
				args.Entries = []LogEntry{}
			}
			reply := AppendEntriesReply{}
			rf.mu.Unlock()
			ok := rf.peers[i].Call("Raft.AppendEntries", &args, &reply)
			if ok {
				rf.mu.Lock()
				defer rf.mu.Unlock()
				if rf.state != Leader {
					return
				}
				if reply.Term > rf.currentTerm {
					rf.currentTerm = reply.Term
					rf.state = Follower
					rf.votedFor = -1
					rf.grantedVote = 0
					rf.persist()
					rf.electionTimer.Reset(ElectionTimeout())
					return
				}
				if reply.Success {
					rf.matchIndex[i] = args.LastLogIndex + len(args.Entries)
					rf.nextIndex[i] = rf.matchIndex[i] + 1
					rf.commitIndex = rf.lastIncludedIndex
					for j := rf.getLastIndex(); j >= rf.lastIncludedIndex+1; j-- {
						sum := 1
						for idx := 0; idx < len(rf.peers); idx++ {
							if idx != rf.me {
								if rf.matchIndex[idx] >= j {
									sum++
								}
							}
						}
						if sum > len(rf.peers)/2 && rf.restoreLogTerm(j) == rf.currentTerm {
							rf.commitIndex = j
							break
						}
					}
				} else {
					if reply.UpToIdx != -1 {
						rf.nextIndex[i] = reply.UpToIdx
					}
				}
			}
		}(i)
	}
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	if args.Term < rf.currentTerm {
		reply.Success = false
		reply.Term = rf.currentTerm
		reply.UpToIdx = -1
		rf.mu.Unlock()
		return
	}
	reply.Success = true
	reply.Term = args.Term
	reply.UpToIdx = -1
	rf.state = Follower
	rf.currentTerm = args.Term
	rf.votedFor = -1
	rf.grantedVote = 0
	rf.persist()
	rf.electionTimer.Reset(ElectionTimeout())
	if args.LastLogIndex < rf.lastIncludedIndex {
		reply.Success = false
		reply.UpToIdx = rf.getLastIndex() + 1
		rf.mu.Unlock()
		return
	}
	if args.LastLogIndex > rf.getLastIndex() {
		reply.Success = false
		reply.UpToIdx = rf.getLastIndex()
		rf.mu.Unlock()
		return
	} else {
		if rf.restoreLogTerm(args.LastLogIndex) != args.LastLogTerm {
			reply.Success = false
			tempTerm := rf.restoreLogTerm(args.LastLogIndex)
			for index := args.LastLogIndex; index >= rf.lastIncludedIndex; index-- {
				if rf.restoreLogTerm(index) != tempTerm {
					reply.UpToIdx = index + 1
					break
				}
			}
			rf.mu.Unlock()
			return
		}
	}
	rf.logs = append(rf.logs[:args.LastLogIndex+1-rf.lastIncludedIndex], args.Entries...)
	rf.persist()
	if args.LeaderCommitIdx > rf.commitIndex {
		rf.commitIndex = Min(args.LeaderCommitIdx, rf.getLastIndex())
	}
	rf.mu.Unlock()
}

func (rf *Raft) committer() {
	for rf.killed() == false {
		time.Sleep(ApplyInterval)
		rf.mu.Lock()
		if rf.lastApplied >= rf.commitIndex {
			rf.mu.Unlock()
			continue
		}
		appliedMsgs := make([]ApplyMsg, 0)
		for rf.commitIndex > rf.lastApplied && rf.lastApplied < rf.getLastIndex() {
			rf.lastApplied++
			appliedMsg := ApplyMsg{
				CommandValid:  true,
				SnapshotValid: false,
				Command:       rf.restoreLog(rf.lastApplied).Command,
				CommandIndex:  rf.lastApplied,
			}
			appliedMsgs = append(appliedMsgs, appliedMsg)
		}
		rf.mu.Unlock()
		for _, appliedMsg := range appliedMsgs {
			rf.applyCh <- appliedMsg
		}
	}
}

func (rf *Raft) persist() {
	data := rf.persistData()
	rf.persister.SaveRaftState(data)
}

func (rf *Raft) persistData() []byte {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	err := e.Encode(rf.currentTerm)
	if err != nil {
		return nil
	}
	err = e.Encode(rf.votedFor)
	if err != nil {
		return nil
	}
	err = e.Encode(rf.logs)
	if err != nil {
		return nil
	}
	err = e.Encode(rf.lastIncludedIndex)
	if err != nil {
		return nil
	}
	err = e.Encode(rf.lastIncludedTerm)
	if err != nil {
		return nil
	}
	data := w.Bytes()
	return data
}

func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 {
		return
	}
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	var currentTerm int
	var votedFor int
	var logs []LogEntry
	var lastIncludedIndex int
	var lastIncludedTerm int
	if d.Decode(&currentTerm) != nil ||
		d.Decode(&votedFor) != nil ||
		d.Decode(&logs) != nil ||
		d.Decode(&lastIncludedIndex) != nil ||
		d.Decode(&lastIncludedTerm) != nil {
		panic("readPersist error")
	} else {
		rf.currentTerm = currentTerm
		rf.votedFor = votedFor
		rf.logs = logs
		rf.lastIncludedIndex = lastIncludedIndex
		rf.lastIncludedTerm = lastIncludedTerm
	}
}

func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {
	return true
}

func (rf *Raft) Snapshot(index int, snapshot []byte) {
	if rf.killed() {
		return
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if index <= rf.lastIncludedIndex || index > rf.commitIndex {
		return
	}
	snapshotLog := make([]LogEntry, 1)
	for i := index + 1; i <= rf.getLastIndex(); i++ {
		snapshotLog = append(snapshotLog, rf.restoreLog(i))
	}
	if index == rf.getLastIndex()+1 {
		rf.lastIncludedTerm = rf.getLastTerm()
	} else {
		rf.lastIncludedTerm = rf.restoreLogTerm(index)
	}
	rf.logs = snapshotLog
	rf.lastIncludedIndex = index
	rf.commitIndex = Max(rf.commitIndex, index)
	rf.lastApplied = Max(rf.lastApplied, index)
	rf.persister.SaveStateAndSnapshot(rf.persistData(), snapshot)
}

func (rf *Raft) leaderSendSnapshot(server int) {
	rf.mu.Lock()
	args := InstallSnapshotArgs{
		Term:              rf.currentTerm,
		LeaderId:          rf.me,
		LastIncludedIndex: rf.lastIncludedIndex,
		LastIncludedTerm:  rf.lastIncludedTerm,
		Data:              rf.persister.ReadSnapshot(),
	}
	reply := InstallSnapshotReply{}
	rf.mu.Unlock()
	ok := rf.peers[server].Call("Raft.InstallSnapshot", &args, &reply)
	if ok {
		rf.mu.Lock()
		defer rf.mu.Unlock()
		if rf.state != Leader || rf.currentTerm != args.Term {
			return
		}
		if reply.Term > rf.currentTerm {
			rf.currentTerm = reply.Term
			rf.state = Follower
			rf.votedFor = -1
			rf.persist()
			rf.electionTimer.Reset(ElectionTimeout())
			return
		}
		rf.matchIndex[server] = rf.lastIncludedIndex
		rf.nextIndex[server] = rf.lastIncludedIndex + 1
	}
}

func (rf *Raft) InstallSnapshot(args *InstallSnapshotArgs, reply *InstallSnapshotReply) {
	rf.mu.Lock()
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		rf.mu.Unlock()
		return
	}
	rf.currentTerm = args.Term
	reply.Term = rf.currentTerm
	rf.state = Follower
	rf.votedFor = -1
	rf.grantedVote = 0
	rf.persist()
	rf.electionTimer.Reset(ElectionTimeout())
	if rf.lastIncludedIndex >= args.LastIncludedIndex {
		rf.mu.Unlock()
		return
	}
	index := args.LastIncludedIndex
	tempLogs := make([]LogEntry, 1)
	for i := index + 1; i <= rf.getLastIndex(); i++ {
		tempLogs = append(tempLogs, rf.restoreLog(i))
	}
	rf.logs = tempLogs
	rf.lastIncludedIndex = args.LastIncludedIndex
	rf.lastIncludedTerm = args.LastIncludedTerm
	rf.commitIndex = Max(rf.commitIndex, rf.lastIncludedIndex)
	rf.lastApplied = Max(rf.lastApplied, rf.lastIncludedIndex)
	rf.persister.SaveStateAndSnapshot(rf.persistData(), args.Data)
	msg := ApplyMsg{
		SnapshotValid: true,
		Snapshot:      args.Data,
		SnapshotTerm:  rf.lastIncludedTerm,
		SnapshotIndex: rf.lastIncludedIndex,
	}
	rf.mu.Unlock()
	rf.applyCh <- msg
}

func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.killed() || rf.state != Leader {
		return -1, -1, false
	}
	index := rf.getLastIndex() + 1
	term := rf.currentTerm
	rf.logs = append(rf.logs, LogEntry{
		Term:    term,
		Command: command,
	})
	rf.persist()
	rf.newLog++
	return index, term, true
}

func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.currentTerm, rf.state == Leader
}

func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{
		peers:             peers,
		persister:         persister,
		me:                me,
		dead:              0,
		applyCh:           applyCh,
		state:             Follower,
		currentTerm:       0,
		votedFor:          -1,
		logs:              make([]LogEntry, 1),
		grantedVote:       0,
		newLog:            0,
		electionTimer:     time.NewTimer(ElectionTimeout()),
		matchIndex:        make([]int, len(peers)),
		nextIndex:         make([]int, len(peers)),
		commitIndex:       0,
		lastApplied:       0,
		lastIncludedIndex: 0,
		lastIncludedTerm:  0,
	}
	rf.readPersist(persister.ReadRaftState())
	if rf.lastIncludedIndex > 0 {
		rf.lastApplied = rf.lastIncludedIndex
	}
	go rf.ticker()
	go rf.applier()
	go rf.committer()
	return rf
}
