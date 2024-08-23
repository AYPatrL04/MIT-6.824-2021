package raft

import "log"

// Debugging
const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (rf *Raft) getLastIndex() int {
	return len(rf.logs) - 1 + rf.lastIncludedIndex
}

func (rf *Raft) getLastTerm() int {
	if len(rf.logs)-1 == 0 {
		return rf.lastIncludedTerm
	} else {
		return rf.logs[len(rf.logs)-1].Term
	}
}

func (rf *Raft) getLastLogInfo(server int) (int, int) {
	newEntryBeginIndex := rf.nextIndex[server] - 1
	lastIndex := rf.getLastIndex()
	if newEntryBeginIndex == lastIndex+1 {
		newEntryBeginIndex = lastIndex
	}
	return newEntryBeginIndex, rf.restoreLogTerm(newEntryBeginIndex)
}

func (rf *Raft) restoreLogTerm(index int) int {
	if index == rf.lastIncludedIndex {
		return rf.lastIncludedTerm
	}
	if len(rf.logs) < index-rf.lastIncludedIndex {
		return -1
	}
	return rf.logs[index-rf.lastIncludedIndex].Term
}

func (rf *Raft) restoreLog(curIndex int) LogEntry {
	if curIndex == rf.lastIncludedIndex {
		return LogEntry{Term: rf.lastIncludedTerm, Command: nil}
	}
	return rf.logs[curIndex-rf.lastIncludedIndex]
}

func (rf *Raft) ValidLog(lastLogTerm, lastLogIndex int) bool {
	lastIdx := rf.getLastIndex()
	lastTerm := rf.getLastTerm()
	return lastLogTerm > lastTerm || (lastLogTerm == lastTerm && lastLogIndex >= lastIdx)
}

func (rf *Raft) GetStateSize() int {
	return rf.persister.RaftStateSize()
}
