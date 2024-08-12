package raft

import (
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

	ApplyInterval       = time.Millisecond * 100
	HBInterval          = time.Millisecond * 200
	ElectionBaseTimeout = time.Millisecond * 800
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

	applyCh  chan ApplyMsg // channel to send applyMsg to service
	state    int
	leaderId int

	currentTerm int
	votedFor    int
	logs        []Entry

	commitIndex int   // index of highest log entry known to be committed
	lastApplied int   // index of highest log entry applied to state machine
	nextIndex   []int // for each server, index of the next log entry to send to that server
	matchIndex  []int // for each server, index of highest log entry known to be replicated on server

	electionTimer  *time.Timer
	heartBeatTimer *time.Timer
}

type Entry struct {
	Term    int
	Command interface{}
}

func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.currentTerm, rf.state == Leader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
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
	VoteGranted  bool
	CurrentTerm  int
	LastLogIndex int
}

type AppendEntriesArgs struct {
	Term          int
	LeaderId      int
	LastLogIndex  int
	LastLogTerm   int
	Entries       []Entry
	LastCommitIdx int
}

type AppendEntriesReply struct {
	CurrentTerm int
	Success     bool
	UpToIdx     int
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term < rf.currentTerm ||
		(args.Term == rf.currentTerm && rf.votedFor != -1 && rf.votedFor != args.CandidateId) {
		reply.VoteGranted = false
		reply.CurrentTerm = rf.currentTerm
		reply.LastLogIndex = len(rf.logs) - 1
		return
	}
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.state = Follower
		rf.persist()
	}
	if !rf.ValidLog(args.LastLogTerm, args.LastLogIndex) || rf.votedFor != -1 && rf.votedFor != args.CandidateId && rf.currentTerm == args.Term {
		reply.VoteGranted = false
		reply.CurrentTerm = rf.currentTerm
		reply.LastLogIndex = len(rf.logs) - 1
		return
	}
	rf.votedFor = args.CandidateId
	rf.electionTimer.Reset(ElectionTimeout())
	reply.VoteGranted = true
	reply.CurrentTerm = rf.currentTerm
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
	rf.logs = append(rf.logs, Entry{
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
			rf.StartElection()
			rf.electionTimer.Reset(ElectionTimeout())
			rf.mu.Unlock()
		case <-rf.heartBeatTimer.C:
			rf.mu.Lock()
			if rf.state == Leader {
				rf.BroadcastHeartBeat()
				rf.heartBeatTimer.Reset(HBInterval)
			}
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
	votes := 1
	rf.votedFor = rf.me
	for i := range rf.peers {
		if i != rf.me {
			go func(i int) {
				reply := RequestVoteReply{}
				rf.sendRequestVote(i, &args, &reply)
				rf.mu.Lock()
				defer rf.mu.Unlock()
				if reply.CurrentTerm > args.Term {
					if reply.CurrentTerm > rf.currentTerm {
						rf.currentTerm = reply.CurrentTerm
					}
					rf.state = Follower
					rf.votedFor = -1
					rf.persist()
					return
				}
				if reply.VoteGranted {
					votes++
					if votes > len(rf.peers)/2 {
						rf.state = Leader
						rf.matchIndex = make([]int, len(rf.peers))
						rf.matchIndex[rf.me] = len(rf.logs) - 1
						rf.nextIndex = make([]int, len(rf.peers))
						for i := range rf.nextIndex {
							rf.nextIndex[i] = len(rf.logs)
						}
						rf.BroadcastHeartBeat()
						rf.heartBeatTimer.Reset(HBInterval)
					}
				}
			}(i)
		}
	}
}

func (rf *Raft) BroadcastHeartBeat() {
	args := &AppendEntriesArgs{
		Term:     rf.currentTerm,
		LeaderId: rf.me,
	}
	for i := range rf.peers {
		if i == rf.me {
			rf.electionTimer.Reset(ElectionTimeout())
			continue
		}
		go func(i int) {
			reply := &AppendEntriesReply{}
			rf.peers[i].Call("Raft.ReceiveHeartBeat", args, reply)
			rf.mu.Lock()
			defer rf.mu.Unlock()
			if reply.CurrentTerm > rf.currentTerm {
				rf.currentTerm = reply.CurrentTerm
				rf.state = Follower
				rf.votedFor = -1
				rf.persist()
				return
			}
		}(i)
	}
}

func (rf *Raft) ReceiveHeartBeat(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term < rf.currentTerm {
		reply.Success = false
		reply.CurrentTerm = rf.currentTerm
		return
	}
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.state = Follower
		rf.persist()
	}
	rf.electionTimer.Reset(ElectionTimeout())
	rf.leaderId = args.LeaderId
	reply.Success = true
	reply.CurrentTerm = rf.currentTerm
}

func (rf *Raft) applier() {
	for rf.killed() == false {
		time.Sleep(ApplyInterval)
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
				Term:          rf.currentTerm,
				LeaderId:      rf.me,
				LastLogIndex:  rf.matchIndex[i],
				LastLogTerm:   lastLogTerm,
				LastCommitIdx: rf.commitIndex,
			}
			if rf.nextIndex[i] < len(rf.logs) {
				if rf.nextIndex[i] > rf.matchIndex[i] {
					args.Entries = append([]Entry(nil), rf.logs[rf.nextIndex[i]:]...)
				} else {
					args.Entries = append([]Entry(nil), rf.logs[rf.matchIndex[i]+1:]...)
				}
			}
			if lastLogTerm < rf.currentTerm {
				args.Entries = append([]Entry(nil), rf.logs[1:]...)
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
				if reply.CurrentTerm > rf.currentTerm {
					rf.currentTerm = reply.CurrentTerm
					rf.state = Follower
					rf.votedFor = -1
					rf.mu.Unlock()
					return
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
										var rep int
										rf.peers[idx].Call("Raft.ReceiveCommit", &rf.commitIndex, &rep)
									}(idx)
								}
							}
							break
						}
					}
					rf.mu.Unlock()
				} else {
					rf.mu.Lock()
					if reply.UpToIdx != -1 {
						rf.nextIndex[i] = reply.UpToIdx + 1
					}
					if reply.CurrentTerm > rf.currentTerm {
						rf.currentTerm = reply.CurrentTerm
						rf.state = Follower
						rf.votedFor = -1
					}
					rf.mu.Unlock()
				}
			}
		}(i)
	}
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term < rf.currentTerm {
		reply.Success = false
		reply.CurrentTerm = rf.currentTerm
		reply.UpToIdx = -1
		return
	}
	reply.Success = true
	reply.CurrentTerm = rf.currentTerm
	reply.UpToIdx = -1
	rf.state = Follower
	rf.currentTerm = args.Term
	rf.votedFor = -1
	if args.LastLogIndex > len(rf.logs)-1 {
		reply.Success = false
		reply.UpToIdx = args.LastLogIndex
	} else {
		if len(args.Entries) > 0 {
			var tmpEntry []Entry
			copy(tmpEntry, rf.logs[args.LastLogIndex:])
			conflictIdx := rf.findConflictIndex(tmpEntry, args.Entries)
			if conflictIdx+args.LastLogIndex+1 < len(rf.logs) {
				rf.logs = rf.logs[:conflictIdx+args.LastLogIndex+1]
			}
			if args.LastLogIndex == 0 {
				if len(args.Entries) > 0 {
					rf.logs = append(rf.logs, args.Entries...)
				} else {
					rf.logs = append(rf.logs[1:], args.Entries...)
				}
			} else {
				rf.logs = append(rf.logs, args.Entries[conflictIdx:]...)
			}
		}
		reply.UpToIdx = len(rf.logs) - 1
	}
}

func (rf *Raft) findConflictIndex(logs []Entry, entries []Entry) int {
	i := 0
	for i < len(logs) && i < len(entries) {
		if logs[i].Term != entries[i].Term {
			return i
		}
		i++
	}
	return i
}

func (rf *Raft) committer() {
	for rf.killed() == false {
		time.Sleep(ApplyInterval)
		var appliedMsg []ApplyMsg
		rf.mu.Lock()
		for rf.commitIndex > rf.lastApplied && rf.lastApplied < len(rf.logs)-1 {
			rf.lastApplied++
			appliedMsg = append(appliedMsg, ApplyMsg{
				CommandValid: true,
				Command:      rf.logs[rf.lastApplied].Command,
				CommandIndex: rf.lastApplied,
			})
		}
		rf.mu.Unlock()
		for _, msg := range appliedMsg {
			rf.applyCh <- msg
			//fmt.Printf("Server %d applied %v\n", rf.me, msg)
		}
	}
}

func (rf *Raft) ReceiveCommit(args *int, reply *int) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if *args > rf.commitIndex {
		rf.commitIndex = *args
	}
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
		logs:           make([]Entry, 1),
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
