# MIT 6.824 (2021) Notes

Personal notes and lab implementations of [MIT 6.824: Distributed Systems](https://pdos.csail.mit.edu/6.824/).

## LECS

Notes see [/Note.md](/Note.md)

- [x] Lec 01
- [x] Lec 02
- [x] Lec 03
- [x] Lec 04
- [x] Lec 05
- [x] Lec 06 (Lab1 Q&A)
- [x] Lec 07
- [x] Lec 08 (Lab2A&2B Q&A)
- [x] Lec 09
- [x] Lec 10
- [x] Lec 11
- [x] Lec 12
- [x] Lec 13
- [x] Lec 14
- [x] Lec 15
- [x] Lec 16
- [x] Lec 17
- [ ] Lec 18
- [ ] Lec 19
- [ ] Lec 20
- [ ] Lec 21

## LABS

Tips see [/LabTips](/LabTips), Implementations see [/src](/src).

- [x] Lab 1
- [x] Lab 2A
- [x] Lab 2B
- [x] Lab 2C
- [x] Lab 2D
- [x] Lab 3A
- [x] Lab 3B
- [x] Lab 4A
- [x] Lab 4B

### Lab1 (MapReduce)
An implementation of basic **MapReduce** framework composed of a master and multiple workers, which communicate with each other through RPC. The master assigns tasks to workers and collects the results. The worker performs the assigned tasks and reports the results to the master.

![MapReduce](/images/MapReduce_Simplify.png)

### Lab2 (Raft)
An implementation of basic **Raft** algorithm. The Raft algorithm is a consensus algorithm that can ensure the consistency of distributed systems. It is mainly used to ensure the Consistency and Partition tolerance (CP in CAP) in distributed systems. The Raft algorithm is divided into three parts: leader election (failover), log replication, and safety. The leader election part is to elect a leader from multiple servers. The log replication part is to ensure that the logs of majority servers are consistent. The safety part is to ensure that the system can only have one leader at a time and prevent the split-brain problems.

![Raft](/images/Raft_Simplify.png)

<h6>The implementation shown above might occur some problems in the following situation:</h6>

<h6>F1 has failed, L1 suddenly failed after committed a log, and as well as F3 and F4 that received the log (F3 and F4 did not commit it yet due to log delay). However, F2 for some reason failed to receive the log and still connecting, and has equal term and longer logs of its own. Then F2 start an election, and F1, F3 reconnected and voted for F2. Therefore, F2 becomes L2, and its local log will be different from the log committed by L1, and cause problems.</h6>

<h6>I did not fix it, but I think it can be fixed with more strict vote rules and additional mechanisms, like adding log time to the entries and compare it before voting. This might not be the best solution, and there should have other mechanisms or ideas to ensure the safety.</h6>

### Lab3 (KVRaft)
An implementation of clients and servers that communicate with each other through RPC. The client sends requests to the server, and the server processes the requests and returns the results to the client. The server uses the Raft to ensure the consistency of the data, and the data that transfers between the client and the server are in the form of key-value pairs, which is similar to the idea of MapReduce. You might simply take the client as the master, the server as the worker, the Put/Append operations as the Map operations, and the Get operations as the Reduce operations, which are assigned by the client (master / coordinator) and processed by the servers (workers) using the Raft algorithm.

![KVRaft](/images/KVRaft_Simplify.png)

<h6>The communication is in reality directly happens between servers. The raft library is in the servers, here is for better understanding.</h6>

### Lab4 (ShardController / ShardKV)
A new concept was introduced here called **Shard**. The data is split to multiple shards, and each shard is assigned to a group. The group is responsible for processing the requests of the shard. The group is composed of multiple servers, using the Raft algorithm to ensure the consistency of the data in the group. Here Shard Controller managing the modifications of the shards and the groups, like adding / deleting / moving shards / servers, and is also responsible for the load balancing of the groups to prevent the hotspots.

![ShardKV](/images/ShardKV_Simplify.png)
