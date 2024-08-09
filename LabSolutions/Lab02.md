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
  - Most possible situation here is Followers commited logs earlier than Leader, in which case if the majority of followers disconnected, the leader will not be able to commit the logs as Followers had done. This can be solved by adding a `MaxCommitIndex` to the `Raft` struct, and call it in Followers only when Leader reached the restriction to commit new logs, to update the `CommitIndex` in Followers.
- Test (2B): concurrent Start()s
  - No problems occurred in tests.
- Test (2B): rejoin of partitioned leader
  - The committed logs can no longer be overwritten by the leader. To prevent such case, the server suffered from network partition should become Follower and update its logs first.
- Test (2B): leader backs up quickly over incorrect follower logs
  - No problems occurred in tests.
- Test (2B): RPC counts aren't too high
  - Make the election timeout and heartbeat timeout longer to reduce the RPC counts.