# Tips for Lab 04

## 4A

This part is quite similar to Lab 3. However, remember to do the load balancing here to prevent a single server from being overwhelmed.

The address is used in the go language to initialize some data structures when directly assigning them to another ready-made data structure. Therefore, in order to copy them completely and not affect the values of the original structure, it is assigned one by one according to the basic data type.

## 4B

This is the most complete and complex part of all labs, it is a combination of all previous parts.

Before start, you might need to go through all the previous parts and make sure you understand them.

For each operation, the client will firstly transfer the `key` to the corresponding shard, which was later used to find the group that responsible for the shard. Then the client will iterate through the group to find the leader, and send the operation.

The leader will validate the key, and send the operation to the raft to update the operation log of itself and followers. After the operation is committed, the `ApplyMsgHandler` will catch the committed operation and apply it to the state machine.

`ck.config = ck.sm.Query(-1)` refers to querying the newest configuration and replacing the current old configuration with it.
