# Solution for Lab 02

## 2A

There are 4 cases that needed to reset the election time to prevent `warning: term changed even though there were no failures`

1. Voting for a candidate
2. Receiving a valid heartbeat
3. Starting an election
4. Leaders sending heartbeats and iterating to itself

## 2B

- Test (2B): basic agreement
  - Simply append logs and send it back by adding it to the `ApplyChan` struct.
- Test (2B): RPC byte count
  - Update the log with restrictions to avoid unnecessary data transfer.
- Test (2B): agreement despite follower disconnection
  - The situation here is a follower disconnected and reconnected, while during the disconnection, the leader has appended logs or even changed. The follower will restart the election due to its timer timeout, and the leader will send the newest log index to transfer it to the follower.
- Test (2B): no agreement if too many followers disconnect
  - Most possible situation here is Followers commited logs earlier than Leader, in which case if the majority of followers disconnected, the leader will not be able to commit the logs as Followers had done. This can be solved by adding a `CommitIndex` to the `Raft` struct, and broadcast it to Followers only when Leader reached the restriction to commit new logs, to update the `CommitIndex` in Followers.
  - Another solution is Leader set its `CommitIndex` after the majority of Followers have received the logs, and send it through the `AppendEntries` RPC in later heartbeats / append logs. This can reduce the number of RPCs, and make sure that logs of Followers is always updated slower than Leader.
- Test (2B): concurrent Start()s
  - Use mutex at every corner that you think it is necessary.
- Test (2B): rejoin of partitioned leader
  - The committed logs can no longer be overwritten by the leader. To prevent such case, the server suffered from network partition should become Follower and update its logs first.
- Test (2B): leader backs up quickly over incorrect follower logs
  - The previous Leader when reconnected and had not updated its state yet while received the log from Clients, it should first renew its state and then decide whether to append the logs or not. It will fail sometimes due to the possible crash of the random timeouts, just pray for that.
- Test (2B): RPC counts aren't too high
  - Make the timeouts longer or use the mentioned solution in the `no agreement if too many followers disconnect` test.

## 2C

Here we only need to fill the `rf.persist()` and `rf.readPersist()` functions. The `rf.persist()` function is called whenever the Raft state changes and needs to be saved to stable storage, and `rf.readPersist()` is called when the Raft restarts and needs to read the persisted state.

The `rf.persist()` function should save:

- `currentTerm`
- `votedFor`
- `logs`

for labs before 2D, and for 2D, we need to save:

- `lastIncludedIndex`
- `lastIncludedTerm`

And the above values should be decoded and read in `rf.readPersist()`.

For `figure8 & figure8(unreliable)`, make sure that the `commitIndex` is updated only when the `term` of the Command is the same as the logged term, and newer than the `commitIndex`. After that, pray for it.

## 2D

The function `rf.CondInstallSnapshot` is not necessary to be implemented if the sychronization is done and locked correctly.

The `rf.Snapshot` here is called by the tester every 10 commands, we only need to check if the logs are updated, and take a newest snapshot of the updated logs.

Here another function `rf.leaderSendSnapshot` is called only when the leader has a newer snapshot than the follower, and the follower will update its logs with the leader's snapshot.

Also be mindful of the various indices and terms of the logs, and the `lastIncludedIndex` and `lastIncludedTerm` of the snapshot.