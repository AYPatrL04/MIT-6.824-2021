# Tips for Lab 03

The tasks here mainly focus on the interaction between Client and Leader. The first thing to do is to understand the whole process of the interaction.

To begin with, Client randomly chooses a Server and guesses that it might be Leader, and retries if not until it finds the Leader. Then, it sends a request to the Leader like `get` or `put/append`. The Leader will process the request through the implemented `Raft` library, sync and wait till everything mentioned in Lab2 has done, and send the result back to the Client. The Client will then receive the result.

## 3A

### Client:

A `seqId` was used here to make sure that the Client can distinguish the responses from different requests. The `seqId` is incremented by 1 each time a request is sent. The Client will wait for the response with the same `seqId` as the request. If the response is not received within a certain time, the Client will retry the request.

### Leader / Server:

For a KVServer, the Leader will process the request and send the result back to the Client. The Leader will also update the `commitIndex` and `lastApplied` to the latest `logIndex` after the request is processed.

```go
clientSeqMap map[int64]int     // map client id to the last seq id	
kvStore      map[string]string // key-value store
applyChMap   map[int]chan Op   // map log index to apply channel handled from raft
```

### Raft:

A problem usually occurring here that the `func TestCount2B(t *testing.T)` always clashes with the `func TestSpeed3A(t *testing.T)`, as the previous one requires low RPC counts per second (about 30 RPCs), while the latter one requires 3 operations per 100ms.

For my original implementation, the RPC was sent every heartbeat (Interval: 35ms), and logs were synced then. However, this will result in a too low efficiency for 3A ( $\frac{1000ms}{35ms/ops}=28.57ops<33.3ops$ ). The maximum acceptable heartbeat interval here is 33ms, while for 2B, the minimum acceptable heartbeat interval is also 33ms. Though by set the interval to 33ms can pass the both tests thanks to the possible accurate lossy of floating-point numbers, it is not an acceptable solution.

Using a channel might help here, and surely I have tried it. But there is a simpler solution that can pass the 2 tests:

```go
const (
    HBInterval = time.Millisecond * 15
)
func (rf *Raft) applier() {
	retry := 0 // retry times
	for rf.killed() == false {
		rf.mu.Lock()
        // directly do apply when there is a new log, check every 15ms
        // otherwise, just wait for 60ms for the next heartbeat
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
	rf.newLog++ // update if there is a new log
	return index, term, true
}
```

and its result (as examples):

```
Test (2B): RPC counts aren't too high ...
  ... Passed --  66 RPCs in total (far fewer than the requirement, regular 110~120 RPCs)

Test: ops complete fast enough (3A) ...
  ... Passed --  20.5 seconds (about 20 ms/op)
```

## 3B

Similar to 2C/2D, just implement the Persisting and Decoding functions. Might need to scan through the whole procedure to prevent deadlocks.

The possible errors occurred during testing might can be fixed by making all operations strictly linearizable.
