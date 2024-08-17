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

	ApplyInterval       = time.Millisecond * 15
	HBInterval          = time.Millisecond * 35
	ElectionBaseTimeout = time.Millisecond * 100
)

func ElectionTimeout() time.Duration {
	return ElectionBaseTimeout + time.Duration(rand.Int63())%ElectionBaseTimeout
}

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	applyCh chan ApplyMsg // channel to send applyMsg to service
	state   int

	currentTerm int
	votedFor    int
	logs        []LogEntry
	grantedVote int

	commitIndex int   // index of highest log entry known to be committed
	lastApplied int   // index of highest log entry applied to state machine
	nextIndex   []int // for each server, index of the next log entry to send to that server
	matchIndex  []int // for each server, index of highest log entry known to be replicated on server

	lastIncludedIndex int
	lastIncludedTerm  int

	electionTimer  *time.Timer
	heartBeatTimer *time.Timer
}

type LogEntry struct {
	Term    int
	Command interface{}
}

func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.currentTerm, rf.state == Leader
}

func (rf *Raft) persist() {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.logs)
	e.Encode(rf.lastIncludedIndex)
	e.Encode(rf.lastIncludedTerm)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
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

// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

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
	if !rf.ValidLog(args.LastLogTerm, args.LastLogIndex) || rf.votedFor != -1 && rf.votedFor != args.CandidateId && rf.currentTerm == args.Term {
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

func (rf *Raft) ValidLog(lastLogTerm, lastLogIndex int) bool {
	lastIdx := len(rf.logs) - 1
	lastTerm := rf.logs[lastIdx].Term
	return lastLogTerm > lastTerm || (lastLogTerm == lastTerm && lastLogIndex >= lastIdx)
}

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.killed() {
		return -1, -1, false
	}
	if rf.state != Leader {
		return -1, -1, false
	}
	index := len(rf.logs)
	term := rf.currentTerm
	rf.logs = append(rf.logs, LogEntry{
		Term:    term,
		Command: command,
	})
	rf.persist()
	return index, term, true
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	for rf.killed() == false {
		select {
		case <-rf.electionTimer.C:
			rf.mu.Lock()
			rf.state = Candidate
			rf.currentTerm++
			rf.persist()
			rf.StartElection()
			rf.electionTimer.Reset(ElectionTimeout())
			rf.mu.Unlock()
		}
	}
}

func (rf *Raft) StartElection() {
	args := RequestVoteArgs{
		LastLogTerm:  rf.logs[len(rf.logs)-1].Term,
		LastLogIndex: len(rf.logs) - 1,
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
	}
	rf.grantedVote = 1
	rf.votedFor = rf.me
	rf.persist()
	for i := range rf.peers {
		if i != rf.me {
			go func(i int) {
				reply := RequestVoteReply{}
				rf.sendRequestVote(i, &args, &reply)
				rf.mu.Lock()
				defer rf.mu.Unlock()
				if rf.state != Candidate || rf.currentTerm > args.Term {
					return
				}
				if reply.Term > args.Term {
					if rf.currentTerm < reply.Term {
						rf.currentTerm = reply.Term
					}
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
						rf.matchIndex[rf.me] = len(rf.logs) - 1
						rf.nextIndex = make([]int, len(rf.peers))
						for i := range rf.nextIndex {
							rf.nextIndex[i] = len(rf.logs)
						}
					}
				}
			}(i)
		}
	}
}

func (rf *Raft) applier() {
	for rf.killed() == false {
		time.Sleep(HBInterval)
		rf.mu.Lock()
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
			var lastLogTerm int
			if rf.matchIndex[i] > len(rf.logs)-1 {
				lastLogTerm = -1
			} else {
				lastLogTerm = rf.logs[rf.matchIndex[i]].Term
			}
			args := AppendEntriesArgs{
				Term:            rf.currentTerm,
				LeaderId:        rf.me,
				LastLogIndex:    rf.matchIndex[i],
				LastLogTerm:     lastLogTerm,
				LeaderCommitIdx: rf.commitIndex,
			}
			if rf.nextIndex[i] < len(rf.logs) {
				if rf.nextIndex[i] > rf.matchIndex[i] {
					args.Entries = append([]LogEntry(nil), rf.logs[rf.nextIndex[i]:]...)
				} else {
					args.Entries = append([]LogEntry(nil), rf.logs[rf.matchIndex[i]+1:]...)
				}
			}
			if lastLogTerm < rf.currentTerm {
				args.Entries = append([]LogEntry(nil), rf.logs[1:]...)
				rf.matchIndex[i] = 0
				rf.nextIndex[i] = 0
				args.LastLogIndex = 0
			}
			reply := AppendEntriesReply{}
			nextI := rf.nextIndex[i]
			logLen := len(rf.logs)
			rf.mu.Unlock()
			if nextI == 0 || ((!(args.Entries == nil || len(args.Entries) == 0)) && logLen > nextI) {
				rf.peers[i].Call("Raft.AppendEntries", &args, &reply)
				rf.mu.Lock()
				if rf.state != Leader {
					rf.mu.Unlock()
					return
				}
				if reply.Term > rf.currentTerm {
					rf.currentTerm = reply.Term
					rf.state = Follower
					rf.votedFor = -1
					rf.mu.Unlock()
					rf.persist()
					return
				} else if reply.Term == rf.currentTerm {
					if reply.UpToIdx != -1 {
						rf.nextIndex[i] = reply.UpToIdx + 1
						rf.matchIndex[i] = reply.UpToIdx
					} else {
						rf.nextIndex[i]--
					}
				}
				rf.mu.Unlock()
				if reply.Success {
					rf.mu.Lock()
					rf.matchIndex[i] = args.LastLogIndex + len(args.Entries)
					rf.nextIndex[i] = rf.matchIndex[i] + 1
					for j := len(rf.logs) - 1; j > rf.commitIndex; j-- {
						sum := 1
						for idx := 0; idx < len(rf.peers); idx++ {
							if idx != rf.me {
								if rf.matchIndex[idx] >= j {
									sum++
								}
							}
						}
						if sum > len(rf.peers)/2 {
							rf.commitIndex = j
							for idx := 0; idx < len(rf.peers); idx++ {
								if idx != rf.me {
									go func(idx int) {
										rf.mu.Lock()
										commitIndex := rf.commitIndex
										var rep int
										rf.mu.Unlock()
										rf.peers[idx].Call("Raft.ReceiveCommit", &commitIndex, &rep)
									}(idx)
								}
							}
							break
						}
					}
					rf.mu.Unlock()
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
	if args.LastLogIndex > len(rf.logs)-1 {
		reply.Success = false
		reply.UpToIdx = len(rf.logs)
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
	rf.logs = append(rf.logs[:args.LastLogIndex+1], args.Entries...)
	rf.persist()
	if args.LeaderCommitIdx > rf.commitIndex {
		rf.commitIndex = Min(args.LeaderCommitIdx, len(rf.logs)-1)
	}
	rf.mu.Unlock()
}

func (rf *Raft) restoreLogTerm(index int) int {
	if index == rf.lastIncludedIndex {
		return rf.lastIncludedTerm
	}
	return rf.logs[index-rf.lastIncludedIndex].Term
}

func (rf *Raft) committer() {
	for rf.killed() == false {
		time.Sleep(ApplyInterval)
		rf.mu.Lock()
		for rf.commitIndex > rf.lastApplied && rf.lastApplied < len(rf.logs)-1 {
			rf.lastApplied++
			appliedMsg := ApplyMsg{
				CommandValid: true,
				Command:      rf.logs[rf.lastApplied].Command,
				CommandIndex: rf.lastApplied,
			}
			rf.mu.Unlock()
			rf.applyCh <- appliedMsg
			rf.mu.Lock()
			rf.persist()
		}
		rf.persist()
		rf.mu.Unlock()
	}
}

func (rf *Raft) ReceiveCommit(args *int, reply *int) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if *args > rf.commitIndex {
		rf.commitIndex = *args
		rf.persist()
	}
	*reply = 1
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{
		peers:          peers,
		persister:      persister,
		me:             me,
		dead:           0,
		applyCh:        applyCh,
		state:          Follower,
		currentTerm:    0,
		votedFor:       -1,
		logs:           make([]LogEntry, 1),
		electionTimer:  time.NewTimer(ElectionTimeout()),
		heartBeatTimer: time.NewTimer(HBInterval),
		matchIndex:     make([]int, len(peers)),
		nextIndex:      make([]int, len(peers)),
		commitIndex:    0,
		lastApplied:    0,
	}
	rf.readPersist(persister.ReadRaftState())
	go rf.ticker()
	go rf.applier()
	go rf.committer()
	return rf
}
