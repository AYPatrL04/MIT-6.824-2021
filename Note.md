<h1 align="center">MIT 6.824 Learning Notes</h1>
<h5 align="center">AYPatrL04</h5>
<h5 align="center">AYPatrL04@gmail.com</h5>

- [Introduction](#introduction)
  - [Labs](#labs)
  - [Focus on infrastructure, not application.](#focus-on-infrastructure-not-application)
  - [Main topics:](#main-topics)
  - [Context](#context)
- [Abstract view](#abstract-view)
  - [Fault Tolerance](#fault-tolerance)
    - [Other failures](#other-failures)
- [Threads \& RPC (Go)](#threads--rpc-go)
  - [Thread of execution](#thread-of-execution)
    - [Definition](#definition)
    - [Why threads?](#why-threads)
    - [Thread challenges](#thread-challenges)
    - [Go and challenges](#go-and-challenges)
    - [Crawler](#crawler)
      - [Serial](#serial)
      - [ConcurrentMutex](#concurrentmutex)
      - [ConcurrentChannel](#concurrentchannel)
    - [RPC: Remote Procedure Call](#rpc-remote-procedure-call)
    - [RPC semantics under failures](#rpc-semantics-under-failures)
- [GFS: Google File System](#gfs-google-file-system)
  - [Storage](#storage)
      - [High performance](#high-performance)
      - [Many servers](#many-servers)
      - [Fault tolerance](#fault-tolerance-1)
      - [Replication](#replication)
      - [Strong consistency](#strong-consistency)
  - [Consistency](#consistency)
      - [Concurrency](#concurrency)
      - [Failures](#failures)
  - [GFS](#gfs)
  - [GFS Data Read Procedure](#gfs-data-read-procedure)
  - [GFS Master](#gfs-master)
  - [GFS File Read Procedure](#gfs-file-read-procedure)
  - [GFS File Write Procedure](#gfs-file-write-procedure)
  - [GFS Consistency](#gfs-consistency)
- [Primary/Backup Replication](#primarybackup-replication)
  - [Failures](#failures-1)
  - [Challenge](#challenge)
  - [Dealing with backup operations (2 approaches)](#dealing-with-backup-operations-2-approaches)
  - [Level of operations to replicate (RSM)](#level-of-operations-to-replicate-rsm)
  - [VM-FT: Exploit virtualization](#vm-ft-exploit-virtualization)
  - [Divergence sources](#divergence-sources)
  - [VM-FT interruptions handling](#vm-ft-interruptions-handling)
  - [VM-FT non-deterministic instructions handling](#vm-ft-non-deterministic-instructions-handling)
  - [VM-FT failover handling](#vm-ft-failover-handling)
  - [VM-FT Performance](#vm-ft-performance)
- [Fault Tolerance: Raft (1)](#fault-tolerance-raft-1)
  - [Single point of failure](#single-point-of-failure)
  - [Split-brain](#split-brain)
  - [Majority rule](#majority-rule)
  - [Protocols using quorums](#protocols-using-quorums)
  - [RSM with Raft](#rsm-with-raft)
  - [Log usage of Raft](#log-usage-of-raft)
  - [Log entry of Raft](#log-entry-of-raft)
  - [Election](#election)
  - [Split vote](#split-vote)
  - [Election Timeout](#election-timeout)
  - [Vote storage](#vote-storage)
  - [Log diverge](#log-diverge)
- [Fault Tolerance: Raft (2)](#fault-tolerance-raft-2)
  - [Log catchup](#log-catchup)
  - [Erasing log entries](#erasing-log-entries)
  - [Log catch up quickly](#log-catch-up-quickly)
  - [Persistence](#persistence)
  - [Service Recovery](#service-recovery)
  - [Using Raft](#using-raft)
  - [Linearizability / Strong Consistency](#linearizability--strong-consistency)
- [Zookeeper](#zookeeper)
  - [Zookeeper is a Replicated State Machine](#zookeeper-is-a-replicated-state-machine)
  - [Zookeeper throughput](#zookeeper-throughput)
  - [Linearizability](#linearizability)
  - [ZNode API](#znode-api)
  - [Summary](#summary)
- [Patterns and Hints for Concurrency in Go](#patterns-and-hints-for-concurrency-in-go)
- [Chain Replication](#chain-replication)
  - [Zookeeper lock](#zookeeper-lock)
  - [Approaches to build Replicated State Machines](#approaches-to-build-replicated-state-machines)
    - [1. Run all operations through Raft / Paxos (used in Lab 3)](#1-run-all-operations-through-raft--paxos-used-in-lab-3)
    - [2. Configuration manager + Primary / Backup replication (more common)](#2-configuration-manager--primary--backup-replication-more-common)
  - [Overview](#overview)
  - [Crash](#crash)
    - [Head crashed (Easiest)](#head-crashed-easiest)
    - [Middle crashed (More complex)](#middle-crashed-more-complex)
    - [Tail crashed (Relatively easy)](#tail-crashed-relatively-easy)
  - [Add replica](#add-replica)
  - [Chain Replication vs. Raft](#chain-replication-vs-raft)
  - [Extension for read parallelism](#extension-for-read-parallelism)
- [Frangipani](#frangipani)
  - [Overview](#overview-1)
  - [Use case](#use-case)
  - [Challenges](#challenges)
  - [Cache coherence / consistency](#cache-coherence--consistency)
  - [Protocol](#protocol)
  - [Atomicity](#atomicity)
  - [Crash recovery](#crash-recovery)
  - [Crash scenarios](#crash-scenarios)
  - [Logs and version](#logs-and-version)
- [Distributed Transaction](#distributed-transaction)
  - [Transactions](#transactions)
  - [Concurrency control](#concurrency-control)
  - [2-phase locking (2PL)](#2-phase-locking-2pl)
    - [Deadlock](#deadlock)
  - [2-phase commit (2PC)](#2-phase-commit-2pc)
  - [2PC crash analysis](#2pc-crash-analysis)
- [Spanner](#spanner)
  - [Organization](#organization)
  - [Challenges](#challenges-1)
  - [Read-write transactions](#read-write-transactions)
  - [Read-only transactions](#read-only-transactions)
    - [Correctness](#correctness)
    - [Bad plan](#bad-plan)
    - [Snapshot Isolation](#snapshot-isolation)
  - [Clock drift](#clock-drift)
  - [Clock synchronization](#clock-synchronization)
  - [Summary](#summary-1)
- [FaRM (Fast Remote Memory)](#farm-fast-remote-memory)
  - [Setup](#setup)
  - [FaRM API](#farm-api)
  - [Kernel bypass](#kernel-bypass)
  - [RDMA (Remote Direct Memory Access)](#rdma-remote-direct-memory-access)
  - [Transaction using RDMA](#transaction-using-rdma)
  - [OOC (Optimistic Concurrency Control)](#ooc-optimistic-concurrency-control)
  - [Strict serializability](#strict-serializability)
  - [Summary](#summary-2)
- [Spark](#spark)
  - [Programming model: RDD](#programming-model-rdd)
  - [Fault tolerance](#fault-tolerance-2)
    - [Narrow dependency](#narrow-dependency)
    - [Wide dependency](#wide-dependency)
  - [Iterative: PageRank](#iterative-pagerank)
  - [Summary](#summary-3)
- [Memcached](#memcached)
  - [Website evolution](#website-evolution)
  - [Eventual consistency](#eventual-consistency)
  - [Cache invalidation](#cache-invalidation)
  - [Fault tolerance](#fault-tolerance-3)
  - [High performance](#high-performance-1)
  - [Protecting DB](#protecting-db)
    - [New cluster](#new-cluster)
    - [Thundering herd](#thundering-herd)
    - [Memcached server failure](#memcached-server-failure)
  - [Race](#race)
    - [Stale set](#stale-set)
    - [Cold cluster](#cold-cluster)
    - [Regions (Primary and Backup)](#regions-primary-and-backup)
  - [Summary](#summary-4)
- [Fork Consistency - SUNDR (Secure UNtrusted Data Repository)](#fork-consistency---sundr-secure-untrusted-data-repository)
  - [Setting: Network file system](#setting-network-file-system)
  - [Focus: Integrity](#focus-integrity)
  - [Safety problem example](#safety-problem-example)
  - [Big idea: Signed logs of operations](#big-idea-signed-logs-of-operations)
  - [Fork consistency](#fork-consistency)
  - [Detecting forks](#detecting-forks)
  - [Snapshot per user and version vector](#snapshot-per-user-and-version-vector)
  - [Summary](#summary-5)

<h6 align="center">======= Lec.01 Fri. 02 Aug. 2024 =======</h6>

# Introduction

## Labs
1) MapReduce
2) Replication Using Raft
3) Replicated Key-Value Service
4) Shard Key-Value Service

## Focus on infrastructure, not application.
- Storage
- Computation
- Communication

## Main topics:
- Fault tolerance
  - Availability (replication)
  - Recoverability (logging / transactions, durable storage)
- Consistency
- Performance
  - Throughput
  - Latency
- Implementation

## Context
- Motivation: Multi-hours of terabytes of data processing, computations, web indexing, ...
- Goal: easy for non-experts to use.
- Approach: 
  - map functions + reduce functions => sequential code
  - MapReduce deals with distribution

# Abstract view
**MapReduce**: an idea come from functional programming.
```markdown
     f1     f2    f3
map   |      |     |
    a->1          a->1  --> reduce --> a->2
    b->1    b->1        --> reduce --> b->2
                  c->1  --> reduce --> c->1
```
```go
type Pair struct {
	key   string
	value int
}

func mapper(key string, value string) []Pair {
	// key: document name
	// value: document contents
	words := strings.Fields(value)
	kvs := make([]Pair, 0)
	for _, w := range words {
		kvs = append(kvs, Pair{w, 1})
	}
	return kvs
}

func reducer(key string, values []Pair) int {
	// key: word
	// values: list of word counts
	count := 0
	for _, v := range values {
		if v.key == key {
			count += v.value
		}
	}
	return count
}
```
1. The input file firstly is split into few pieces.
2. The schedulers run `worker` processes on a set of machines, calling `Map` and `Reduce` functions when appropriate.
3. The `master` process assign tasks to the workers.
4. The `Map` workers read the input files, and turn them into a list of key / value pairs (store intermediate key / value files on local disks).
5. The `Reduce` workers read the intermediate files (might need network communications to retrieve files), merge them into lists of values as output.

Expensive: shuffle

## Fault Tolerance
**Basic plan**: Coordinator reruns map / reduce tasks if workers fail to respond.

**Can maps / reduces run twice?** Yes, because they are functional / deterministic.

The different machines running same task will produce the same output and write it in the intermediate files. The different procedures will do atomic rename to make sure that only one procedure writes the final output.

### Other failures

**Will coordinator fail?** It cannot fail.
**Slow workers?** They are called `stragglers`, doing backup tasks. Jobs can be replicated to other workers, therefore the performance is not affected.

<h6 align="center">======= Lec.02 Sat. 03 Aug. 2024 =======</h6>

# Threads & RPC (Go)

- Why Go?
  - Good support for threads / RPC
  - Garbage collection
  - Type safe
  - Simple
  - Compiled

## Thread of execution

### Definition

- A thread is a sequence of instructions that can be executed independently.
- Memories are shared between threads.

For Go, using `go` keyword to start a new thread. The `main` function will wait for all threads to finish.

There are other status like:
- `exit` is usually implicit when a function created by key word `go` returns.
- `stop` when the thread is blocked, the go runtime will stop the thread and work on other threads.
- `resume` pick the status of the stopped thread and put it aside, and load it back to the processor when it is ready.

### Why threads?

- Express concurrency
  - I / O concurrency
  - Multi-core parallelism
  - Convenience

### Thread challenges

- Race conditions
  - Avoid sharing
  - Use locks

Go usually has a Race Detector, using `-race` when running the program. `go run [-race] name.go`

- Coordination
  - Channels or Condition Variables

- Deadlocks
  - Avoid circular waits
  - Use timeouts

### Go and challenges

- 2 plans
  - Plan A: Channels (no sharing)
  - Plan B: Locks + Condition Variables (shared memory)

```go
func requestVote() bool {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	return true
}

func voteCount01() { // can be detected to have race condition with "go run -race xxx.go"
	rand.Seed(time.Now().UnixNano())

	count := 0
	finished := 0

	for i := 0; i < 10; i++ {
		go func() {
			vote := requestVote()
			if vote {
				count++
			}
			finished++
		}()
	}
	for count < 5 && finished != 10 {
		// wait
	}
	if count >= 5 {
		println("received 5+ votes!")
	} else {
		println("lost")
	}
}

func voteCount02() { // a kind of solution using mutex
	rand.Seed(time.Now().UnixNano())

	count := 0
	finished := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		go func() {
			vote := requestVote()
			mu.Lock()
			defer mu.Unlock()
			if vote {
				count++
			}
			finished++
		}()
	}
	for {
		mu.Lock()
		if count >= 5 || finished == 10 {
			break
		}
		mu.Unlock()
	}
	if count >= 5 {
		println("received 5+ votes!")
	} else {
		println("lost")
	}
	mu.Unlock()
}

func voteCount03() { // a kind of solution using mutex and condition variable
	rand.Seed(time.Now().UnixNano())

	count := 0
	finished := 0
	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	for i := 0; i < 10; i++ {
		go func() {
			vote := requestVote()
			mu.Lock()
			defer mu.Unlock()
			if vote {
				count++
			}
			finished++
			cond.Broadcast()
		}()
	}

	mu.Lock()
	for count < 5 && finished != 10 {
		cond.Wait()
	}
	if count >= 5 {
		println("received 5+ votes!")
	} else {
		println("lost")
	}
	mu.Unlock()
}

func voteCount04() { // a kind of solution using channel
	rand.Seed(time.Now().UnixNano())
	count := 0
	finished := 0
	ch := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			ch <- requestVote()
		}()
	}
	for count < 5 && finished < 10 {
		v := <-ch
		if v {
			count += 1
		}
		finished += 1
	}
	if count >= 5 {
		println("received 5+ votes!")
	} else {
		println("lost")
	}
}
```

### Crawler

#### Serial

```go
type Fetcher interface {
    Fetch(url string) (urls []string, err error)
}

type fakeResult struct {
    body string
    urls []string
}

type fakeFetcher map[string]*fakeResult

func (f fakeFetcher) Fetch(url string) ([]string, error) {
    if res, ok := f[url]; ok {
        fmt.Printf("found: %s\n", url)
        return res.urls, nil
    }
    return nil, fmt.Errorf("not found: %s", url)
}

func Serial(url string, fetcher Fetcher, fetched map[string]bool) {
    if fetched[url] {
        return
    }
    fetched[url] = true
    urls, err := fetcher.Fetch(url)
    if err != nil {
        return
    }
    for _, u := range urls {
        Serial(u, fetcher, fetched)
    }
    return
}
```

#### ConcurrentMutex

```go
type fetchState struct {
    mu      sync.Mutex
    fetched map[string]bool
}

type Fetcher interface {
    Fetch(url string) (urls []string, err error)
}

func (f fakeFetcher) Fetch(url string) ([]string, error) {
    if res, ok := f[url]; ok {
        fmt.Printf("found: %s\n", url)
        return res.urls, nil
    }
    return nil, fmt.Errorf("not found: %s", url)
}

type fakeFetcher map[string]*fakeResult

type fakeResult struct {
    body string
    urls []string
}

func ConcurrentMutex(url string, fetcher Fetcher, f *fetchState) {
    f.mu.Lock()
    already := f.fetched[url]
    f.fetched[url] = true
    f.mu.Unlock()
  
    if already {
        return
    }
  
    urls, err := fetcher.Fetch(url)
    if err != nil {
        return
    }
    var done sync.WaitGroup
    for _, u := range urls {
        done.Add(1)
        go func(u string) {
            defer done.Done()
            ConcurrentMutex(u, fetcher, f)
        }(u)
    }
    done.Wait()
}
```

#### ConcurrentChannel

```go
type Fetcher interface {
    Fetch(url string) (urls []string, err error)
}

func (f fakeFetcher) Fetch(url string) ([]string, error) {
    if res, ok := f[url]; ok {
        fmt.Printf("found: %s\n", url)
        return res.urls, nil
    }
    return nil, fmt.Errorf("not found: %s", url)
}

type fakeFetcher map[string]*fakeResult

type fakeResult struct {
    body string
    urls []string
}

func worker(url string, ch chan []string, fetcher Fetcher) {
    urls, err := fetcher.Fetch(url)
    if err != nil {
        ch <- []string{}
    } else {
        ch <- urls
    }
}

func coordinator(ch chan []string, fetcher Fetcher) {
    n := 1 // n represents the number of running workers
    fetched := make(map[string]bool)
    for urls := range ch {
        for _, url := range urls {
            if !fetched[url] {
                fetched[url] = true
                n++
                go worker(url, ch, fetcher)
            }
        }
        n--
        if n == 0 {
            break
        }
    }
}

func ConcurrentChannel(url string, fetcher Fetcher) {
    ch := make(chan []string)
    go func() {
        ch <- []string{url}
    }()
    coordinator(ch, fetcher)
}
```

### RPC: Remote Procedure Call

- Goal: make a function call that is implemented on another machine. E.g. `f(x)` on client get the result from server where `f` is implemented.

```markdown
 _____________        _____________
|             |      |             |
| Client f(x) |      | Server f(x) |
|_____________|      |_____________|
      |                     |
 _____|_______        ______|______
|             |      |             |
| Client Stub |      | Server Stub |
|_____________|      |_____________|
      |_____________________|
```
- stub represents a kind of concept that use a controllable sub system to replace a certain function of the original system.
- the procedure of marshall and unmarshall parameters happens in the stub.
- the stubs shown in the graph are literally the same object.

```go
// RPC
const (
	OK       = "OK"
	ErrNoKey = "ErrNoKey"
)
type Err string
type GetArgs struct {
    Key string
}
type GetReply struct {
    Err   Err
    Value string
}
type PutArgs struct {
    Key   string
    Value string
}
type PutReply struct {
    Err Err
}
```

<table>
  <tr>
    <td>Client</td>
    <td>Server</td>
  </tr>
  <tr>
    <td>
      <pre>
        <code>
func connect() *rpc.Client {
    client, err := rpc.DialHTTP("tcp", "localhost:1234")
    if err != nil {
        log.Fatal("dialing:", err)
    }
    return client
}
func get(key string) string {
    client := connect()
    args := GetArgs{"subject"}
    reply := GetReply{}
    err := client.Call("KV.Get", &args, &reply)
    if err != nil {
        log.Fatal("error:", err)
    }
    client.Close()
    return reply.Value
}
func put(key string, value string) {
    client := connect()
    args := PutArgs{"subject", "6.824"}
    reply := PutReply{}
    err := client.Call("KV.Put", &args, &reply)
    if err != nil {
        log.Fatal("error:", err)
    }
    client.Close()
}
        </code>
      </pre>
    </td>
    <td>
      <pre>
        <code>
func server() {
    kv := new(KV)
    kv.data = make(map[string]string)
    rpcs := rpc.NewServer()
    rpcs.Register(kv)
    l, e := net.Listen("tcp", ":1234")
    if e != nil {
        log.Fatal("listen error:", e)
    }
    go func() {
        for {
            conn, err := l.Accept()
            if err == nil {
                go rpcs.ServeConn(conn)
            } else {
                break
            }
        }
        l.Close()
    }()
}
func (kv *KV) Get(args *GetArgs, reply *GetReply) error {
    kv.mu.Lock()
    defer kv.mu.Unlock()
    val, ok := kv.data[args.Key]
    if ok {
        reply.Err = OK
        reply.Value = val
    } else {
        reply.Err = ErrNoKey
        reply.Value = ""
    }
    return nil
}
func (kv *KV) Put(args *PutArgs, reply *PutReply) error {
    kv.mu.Lock()
    defer kv.mu.Unlock()
    kv.data[args.Key] = args.Value
    reply.Err = OK
    return nil
}
        </code>
      </pre>
    </td>
  </tr>
</table>

### RPC semantics under failures

- **At-least-once** (retry until success)
- **At-most-once** (filter out duplicate requests)
- **Exactly-once** (needed to be built in lab3)

<h6 align="center">======= Lec.03 Sat. 03 Aug. 2024 =======</h6>

# GFS: Google File System

- Storage
- GFS and its design
- Consistency

## Storage

The build of fault-tolerant storage system requires a durable storage system.

The application itself is usually stateless, and the storage holds the persistent state.

#### High performance

- Shard: concurrent access
- Data across servers: hardware limitations like CPU or network bandwidth

#### Many servers

- Constant faults: though the failure rate for single machine is low, it is high for a large number of machines, so the fault tolerance is required.

#### Fault tolerance

- Replication: if one fails, the other can take over.

#### Replication

- Potential inconsistency: the replicas might be inconsistent.

#### Strong consistency

- Lower performance: the system needs to wait for the slowest replica to respond.

## Consistency

An ideal consistency is that the distributed system behaves as if there is only one system working.

#### Concurrency

- W1: write 1, W2: write 2 (Concurrent writes)
- R1: read 1, R2: read 2 (Concurrent reads)
- Regardless of the order of W1 and W2, the results of R1 and R2 should be the same.
- Can be achieved by Distributed Locks.

#### Failures

Usually the replication is used to tolerate failures. However, the amateur replication might cause the readers to read the different versions of the data.

For example, W1 and W2 concurrently write 1 and 2 respectively to both replicas. The readers might read 1 from one replica and 2 from the other replica.

## GFS

High performance, replication, fault tolerance, but hard to maintain consistency.

A mapper can read data from GFS(thousands of disks) at a speed of 10,000 MB/s while in that time a single disk can only read at a maximum speed of 30 MB/s.

- Big: large data set
- Fast: automatic sharding
- global: all apps see same files
- fault-tolerant: automatic

## GFS Data Read Procedure

![GFS Architecture](/images/GFS_Architecture.png)

1. The client sends a request to the master for the location of the data.
2. The master does query work to find the corresponding chunk server.
3. The client retrieves the data from the chunk server.
- Chunk is a part of a certain big file. Here each chunk is 64MB.

## GFS Master

**Works:**

- Maintain the mapping relationship between files and chunks. (Usually stored in memory such that Master can respond quickly)
- Maintain the version of each chunk handle.
- Maintain the list of chunk servers.
  - Primary: the first replica of the chunk.
    - The lease time of the primary server
  - Secondaries: the other replicas of the chunk.
  - Usually the chunks are replicated 3 times, on 3 different servers.
- Log + checkpoints: to recover from failures.

**Which of the data needs to be stored stably (such as in disk)?**
- The mapping relationship between files and chunks? √
  - When crush occurs and the memory is lost, the mapping relationships need to be recovered from the disk, otherwise the files will be lost.
- The version of each chunk handle? √
  - The master needs the version to determine which servers have the latest version of the chunk.
- The list of chunk servers? ×
  - The list of chunk servers can be recovered from the chunks themselves.
- The log? √
  - The log is used to recover the system from failures.

## GFS File Read Procedure

1. The client sends a request to the master for the data located at X file with Y offset.
2. The master returns list of chunk servers, chunk handle, and version number.
3. The client caches the list. 
   - In current design, master has only 1 server. The client caches the list can reduce the load of the master.
4. The client reads the data from the closest chunk server. 
   - Minimize network traffic.
5. The chunk server checks the version number, and returns the data to the client if correct.
   - Avoid reading the outdated data.

## GFS File Write Procedure

<h5>Here mostly focus on the append operation, as it is more common in GFS.</h5>

1. The client sends a request to the master for the location to write the data corresponding to the filename.
2. The master query the mapping relationship between the file and the chunk handle, and the mapping relationship between the chunk handle and the chunk servers, and returns the list of chunk servers(primary, secondaries, version number).
   - Has primary
     - Continue
   - No primary
     - Choose one of the chunk servers as the primary, and the others as the secondaries.
       - Master will update version, store it and send its newest to primary and secondaries, and the lease to the primary. Here primary and secondaries needed to store the version to the disk.
3. The client sends the data to the chunk servers(primary and secondaries). Here client only visit the nearest secondary, and the secondary will forward the data to the next chunk server, till which the data has not been stored yet.
   - This can increase the throughput of the client, and reduce the use of network resources.
4. The client sends a message to the primary to inform the append operation.
   - The primary needs to:
     1. Check the version number, and reject if the version mismatch.
     2. Check the lease, and reject the mutation operation if the lease is expired.
     3. Choose an offset to store the data.
     4. Store the data.
5. The primary sends the message to the secondaries to inform the append operation.
6. The secondaries store the data to the corresponding offset, and send the message back to the primary.
7. The primary sends the message to the client to inform the success of the append operation.

**There are situations that the append operation failed that the primary has stored the data, and a certain secondaries has not. In this case, the Client will receive an error message, and the client will retry the append operation. This is the Do At-Least-Once.**

- If append operation failed, and the client retries the operation, will the offset be the same with the previous?
  - No. Primary will choose a new offset to store the data. Assume a primary and 2 secondaries, the previous operation might be successful for p and s1 while failed for s2, and the retry operation needs to store the data to a new offset, at which it might be successful (situations and solutions vary).

Here the replicates records can be duplicated. For an append operation to a certain data, an ID will be bound to the data, and if same ID is found, the second one will be discarded. Meanwhile, the change of data will be checked by checksums and ensure the data would not be modified.

All the servers are trusted because it is a completely internal system, so there is no permission problem.

## GFS Consistency

To simplify the problem. When an append operation ends, and a read operation starts.

Here, assume the existence of Master(M), Primary(P), and Secondary(S).

**From a certain time point, M failed to get the response of Ping-Pong from P, and M cannot determine whether P is alive or not.**
  1. M has to wait till the lease expires, otherwise there might appear 2 P.
  2. At the time point, some clients might still be interacting with this P.

Assume that we pick a new P, and the old P is still alive. This is called the split-brain problem, which will result in many out-minded problems, such as the confusion of the sequence of the data I/O, and even result in 2 M.

Thus, here M knows the expiration of the lease, and will wait till its expiration before picking a new P to maintain the consistency.

- How to get stronger consistency?
  - Update all P + S or none. (Similar to Transactions)

Google has built other systems to ensure the consistency of the data, such as Spanner. Here GFS is for running MapReduce, and the consistency is not the top priority.

<h6 align="center">======= Lec.04 Tue. 06 Aug. 2024 =======</h6>

# Primary/Backup Replication

- Failures
- Challenge
- 2 approaches
  - State transfer
  - Replicated state machine
- Case study
  - VM Fault Tolerance (VM-FT)

## Failures

Common failure scenarios:

- Fail-Stop failures: infrastructure failure or components of the computer does not work well and eventually stops the computer.
- Logic bugs, configuration errors
- Malicious errors
- Handling stop failure

Here we assume that there are no errors with the system and software, no logical errors or configuration errors, and no attackers. In other word, we only focus on **Handling stop failure**.

## Challenge

- Has primary failed?
  - the machine might fail to visit the primary due to network partition or machine failed, and in this case some strategies should be applied in advance to avoid the co-existence of 2 primaries (assume that there are only 1 primary allowed).
- Split brain system
  - if each partition of networks has its own primary, and clients interact with these primaries, there might be huge difference in data, version, etc., and if the system restart, we have to handle it by ourselves (similar to the handling of git merge).
- Keep primary / backup in sync
  - if primary failed, the backup should be able to handle the work of the primary, which requires the backup to sync to the newest primary when primary updated.
  - apply changes in order
  - avoid non-determinism: same changes should be consistent for primary and backup.
- Failover
  - backup is supposed to take over the work of the primary when the latter one occurs problems, but before switch to backup, it should be guaranteed that primary has finished all works, that is to say, it is not supposed to switch to backup if primary is still chatting with client, which would be even harder when facing network partition situations.

## Dealing with backup operations (2 approaches)

1. **State Transfer**
   - Primary makes checkpoints, and sync the checkpoints to the backup before sending response to the client.
2. **Replicated State Machine (RSM)**
   - Primary receives data from client, and generates the operations that would be synced to and done by backup before the primary sending response to the client.

The similar part of the 2 approaches is that the sync operations are always done before the primary replies to the clients. In this case when primary fails, the backup can take the state that is the same as the primary.

Most of the systems choose the 2nd approach, as it would be more expensive to sync the possible gigabytes of states generated by the operations. It would be much less expensive if backup can do the operations and generate the states itself.

GFS also applied the 2nd approach (RSM). See [GFS File Write Procedure](#gfs-file-write-procedure) Step 4 - 7.

## Level of operations to replicate (RSM)

- Application-level operations 
  - Such as GFS file append or write. If using RSM on application-level operations, the semantics of these operations should also be known by RSM, and it should know how to deal with these operations. <ins>if applying an RSM approach on the kind of application-level, <strong>the Application Itself</strong> should be modified to perform or play part of the RSM approach</ins>
- Machine-level operations
  - The state here refers to the state of registers or memories, and the operations are traditional computer instructions. In this level, the RSM only focus on the most basic machine instructions rather than applications or operating systems.
  - A traditional way of applying machine-level operations is to buy those machines or processors that support replicate / backup themselves, but it is usually expansive.
  - Instead of hardware replication, Virtual Machine can be used here.

## VM-FT: Exploit virtualization

The virtual replication is transparent to applications, and appears to clients as if the system is a single machine **(Strong consistency)**. VMWare is based on this in early versions. The shortcoming is that this supports only single-core rather than multicore. Maybe it is supported by later VM-FT, but the solution might be State Transfer rather that pure RSM.


```markdown
            Primary                        Backup
 _____________________________         ______________
| Applications | Applications |       | Applications |
|______________|______________|       |______________|     storage
|   Linux VM   |   Other VMs  |       |   Linux VM   |   ____________
|______________|______________|       |(same version)|  |  ________  |
  | Virtual Machine Monitor | logging |______________|  | |  flag  | |
  |  (Hypervisor or VM-FT)  | ======> |     VM-FT    |  | |________| |
  |_________________________|         |______________|  |____________|
        |   Hardware   |              |   Hardware   |
        |______________|              |______________|
               |
       ________|____________________________________
                         _____|_____
                        |  Clients  |
                        |___________|
```

When interruption happens, the VM-FT act as Hypervisor will receive the signal of interruption, and it will:

1. send the interruption signal to a backup computer over a logging channel
2. send the interruption signal to the actual VM, like the Linux VM running on guest space as shown in the graph.

Similarly, when client send packets to primary and the hardware of primary occurred interruptions, VM-FT will send the interruption to backup through logging channel, and send it to the current VM. The VM then writes the data to the virtual network card in VM-FT, and the VM-FT will write the data to the physical network card to send off the response to the client. Meanwhile, the backup also receives the interruption send by the primary VM-FT, but it will do nothing when its virtual network card received the data.

Assume that there is a storage server outside and connected to the VM-FTs through network, the storage use a flag to record the state of primary and backup, and who is primary.

When primary and backup occurs the network partition problem while they are still able to communicate with the storage server, they will assume that the opposite machine is down, and trying to act as the new primary to change the flag in storage through the test-and-set (sets a new value and returns the original value) atomic operation, and whoever finished the operation in advance will be marked as new primary. The latter one will receive the changed value and realize that it is the latter one, and give up becoming primary (terminate itself).

- When is the flag initialed?
  - Firstly a primary is started, and a backup is started for backup. Then a repair plan is needed to guarantee that if primary fails, the backup can take its place. In VM-FT, the repair plan is executed manually by monitor software. It creates a new replica based on the VM image, and when the backup started and finished the backup, the flag is reset according to the protocol. 
- What will happen if the logging channel or the channel the client used to access the server breaks?
  - If the channel broke, the server will no longer serve the client as long as primary and backup themselves work well, and nothing can be done except wait until the network being repaired.
- A kind of logging channels were implemented by UDP for efficiency. If it fails like a certain packet was not confirmed, will primary simply think backup fails and performs no redial?
  - No because of the timer interruptions. It triggers like every 10ms, and if there is a packet failed to be received, primary will try to resend it to backup. If it is still occurring problems, the work might be stopped.

## Divergence sources

- Non-deterministic instructions
  - E.g. the instruction to get the time as it is hard to ensure that the primary and backup execute in the same time, and the received values are usually different
- Input packets / timer interruptions
  - E.g. the interrupters might be inserted into the instructions stream between 1st and 2nd instructions of primary while between 2nd and 3rd instructions of backup, which might result in the inconsistent states between primary and backup in following operations. So the interruptions are supposed to be delivered at the same point in the instruction streams.
- Multi-core
  - The solution in the paper of this lecture is not allowing multicore to prevent the possible concurrency. For example, on primary there are 2 threads of different cores racing for the lock, and it should be guaranteed that the thread of backup that eventually wins the lock are the same as the thread of primary, which requires a bunch of machinery or complexities that are hard to deal with. Here assumes that the processor is unique.

## VM-FT interruptions handling

When received the interruption, VM-FT knows how many instructions CPU has executed (e.g. 100 instructions), and calculate a position (e.g. 100), inform the backup that when the instructions to execute the interruption after the 100th instruction. Most of the processors (e.g. x86) will stop at the required instruction and give back control to the operating system (here the VM monitor).

To arrange this, the backup needs to lag behind one message.

**The deterministic operations do not need to sync through logging channel as they have a copy of all the instructions.**

## VM-FT non-deterministic instructions handling

It firstly scans through all the non-deterministic instructions in Linux before starting the guest space or boot to the VM, and makes sure that they are transferred into invalid instructions. When guest space executing these non-deterministic instructions, it transfers control to the hypervisor through the trap, and the hypervisor will emulate to execute the instruction, and record the result to a certain register and send it back. The backup will execute this instruction and trap into kernel as well, and it will wait until primary syncs the result of the instruction, thus achieve the same result on this non-deterministic instruction.

## VM-FT failover handling

E.g. Primary had a counter as 10, and client sent a request `inc` to increase it to 11, and the primary did it but failed right before sending response to the client. If backup handled it now by sync to the instruction from logging channel, the `inc` operation will not be executed by backup. If the client sent the `inc` request again, what it will get is 11 rather that 12.

To prevent the above situations, VM-FT made an **Output Rule**. Before primary sending response to client, it will send message to back up through logging channel. When backup received the message and did the same operations, it will send `ack` message to primary (similar to TCP). Only after primary received the `ack` and ensured that backup can have the same state will it send response to client.

The output rule can be seen in any kind of replication systems (e.g. raft or zookeeper).

**As it is a network request from client, though it is a deterministic instruction, it is needed to inform backup and wait for the `ack` message before respond to client.**

**It is possible for client to receive the response twice.** And it can be dealt with TCP.

## VM-FT Performance

**As most of VM-FT operations based on level of machine instructions or level of interruptions, the performance paid a hit.**

According to the statistics from the paper, the performance keeps a rate of 0.94 ~ 0.99 comparing with non-FT when running at primary / backup situations, while if the network transmit and receive with huge amount of data, the performance declines apparently with a decline rate of nearly 30\%. The reason might be that primary needs to send the data to the backup and wait until backup processed all the data can it sends response to the client.

This explains why people prefers to use RSM on application-level rather than instruction-level. However, they usually need to modify the application if using RSM on application-level like GFS.

**All the works done here is never for higher performance or efficiency, but only for a stronger consistency and partition tolerance. In other word, here only focus on the C and P in the traditional CAP theory.**

<h6 align="center">======= Lec.05 Wed. 07 Aug. 2024 =======</h6>

# Fault Tolerance: Raft (1)

One of the main components in Distributed Replication Protocol.

## Single point of failure

It happens when the **Coordinator(MapReduce) / Master(GFS) / Storage(VM-FT test-and-set)** fails.

In these solutions, the single machine managements were applied rather that multi-instances / multi-machines for the purpose of avoiding the **Split-brain**.

However, in most cases, the single point of failure is acceptable, as the failure rate of a single machine is much lower that multi-machines, and the cost to recover it is lower as well, which only need to wait a short period of time to restart.

## Split-brain

Suppose that there are 2 servers named S1 and S2.

Now here is a client C1 want to become Primary, so it sends test-and-set request to both of the servers. And assume that S1 responses but S2 not due to some problems. Now C1 might just think it becomes the Primary.

Here the reason why S2 made no responses might be:
1. S2 failed / has down. Then there will be no problem that C1 becomes Primary as it would be like the single point situations.
2. The network partition occurred between S2 and C1, and only C1 can not get access to S2. Then if there is another client C2 also sent a request to S1 and S2, though S1 would make a fail response, S2 would be success. If the protocol are not that well-considered, the system will appear 2 Primaries, which is known as Split-brain.

To avoid the Split-brain problem, the network partition should firstly be solved.

## Majority rule

The majority here refers to that of all the machines in the systems, regardless of the state of each machine, if a client sent the test-and-set request to them, and there are more than half of the machines approved, than the client will be considered as Primary.

E.g. 
- 5 Machines, 3 Approved = Success, 2 Approved = Fail.
- 4 Machines, 3 Approved = Success, 2 Approved = Fail.

The majority rule can ensure the existence of the overlap between different clients, thus only 1 client can receive the support of majority.

The Raft is mostly the same as:

- If network partition happened, there will be at most 1 partition has majority, and only this partition can continue to work.
- **If no partition had majority (e.g. 3 -> 1, 1, 1), the whole system will be unable to run.**
- Assume a situation of <math>2x+1</math> machines, then the number of machines that can down at the same time is at most <math>x</math>, as the rest of machines can still reach the number of <math>x+1 > x</math> and become majority.

The client itself also perform as a server and will vote for itself. In Raft, the `candidates` will vote for themselves, and the `leader` will also vote and record itself.

## Protocols using quorums
Since 1990s, widely being used since 2000s:

- Paxos
- View-Stamped replication (VR)

The Raft appears around 2014, and it can be used to implement a **Complete Replicated State Machine**

## RSM with Raft

The systems working normally be like:

- Client sends query request to Leader.
- Leader appends the request log to the end of Raft Sequential log.
- Leader sync the added log records to other K / V machines through network.
- Other machines append the log, respond `ACK` to Leader.
- Leader receives the `ACK`, and Leader sends the operation log to its own K/V Application.
- K / V Application does the K / V query, and send the result to Client after the Leader and `ACK`ed followers reach the majority amount.

When error occurred:

- Client sends request to Leader.
- Leader syncs the log and receive the `ACK`.
- Leader downs when responding to Client.
- **Other machines elect a new Leader.**
- Request time out, Client retries. **As the internal failover happened in the System, Client will request the new Leader.**
- New Leader records the log and syncs it to other machines and get `ACK`.
- New Leader responds to Client.

Here the log of the rest of machines might have **duplicate** requests, and it is required to detect these requests.

Similarly, Clients here are required to have the retry mechanism to prevent the possible log **loss**.

In real system designs, the data will be split to multiple Raft instances (shards) with their own Leaders, thereby averaging the requests load to other machines.

Clients have the access list of all servers. When old Leaders down, Clients will randomly resend the request to someone of these servers until success.

## Log usage of Raft

- Retransmission: The message might fail to be transmitted from Leader to Followers, and the log can help with retransmission.
- Order: The operations need to be synced to all replicas should in same order.
- Persistence: To support retransmissions, recoveries, etc..
- Space tentative: When Followers received the log from Leader, they did not know which operations were committed, and needed to wait for a while until they were committed will they do the following operations. Here some spaces are needed for these tentative operations, which is suitable to use log.

**The logs might be delayed sometimes in some Machines, but they will finally sync to and be identical to other servers.**

## Log entry of Raft

For every log entry:

- Command (received instructions and operations): Ensure the logs are identical
- Leader's term (in current system): Elect leaders periodically

**Uncommitted log entries have the possibility to be overwritten.**

## Election

- Leader occurred network partition with followers.
- Followers reelect as they missed the heartbeats of Leader and the election timer was timeout.
  - When Leader did not receive new messages from Followers, it will still send heartbeat periodically to inform that it is still the Leader. It is a kind of append log, but will not be recorded. The messages carried by these kind of logs carry are various, like the length of the current log or what the last log entry should be like, and thus helps Followers to sync up to date.
- Assume that the election timer of Follower 1 was timeout earliest. It increases its term count by 1, and launches the election, votes for itself, and request for vote from original Leader and other Followers.
- Followers respond to the Follower 1, and the Leader does not due to network partition.
- Follower 1 becomes new Leader.
- Clients do the failover and resend the request to new Leader, and same with the following requests.

**If client still sent the request to the original Leader who recovered the connections with other original Followers:**

- The original Leader receives the request from Client and tries to send log to other Followers.
- The new Leader receives the log, and refuses to append it to its own logs, and replies the original Leader with new term count.
- The original Leader will realize that it is no longer the Leader, and either becomes Follower or relaunch the election, but will not continue to serve as Leader, **thus the Split-brain problem will not happen.**
- Client will receive a failure or refuse to serve, and resend the request to the new Leader.

## Split vote

- The election timers of Follower 1 and Follower 2 are timeout at the same time by coincidence.
- Both of them vote for themselves, and receive a reject from each other, and both failed to become majority or new Leader.
- After the election timeout, they repeat the procedure above, and it might lead to dead loop.

**To prevent such kind of dead loop, the election timeout is usually randomized.**

According to the paper, the timer will be set to a random value from 150ms to 300ms, and there will be a time when the timer of one Follower has up while the other does not, and the Follower launches a new election, reaches the majority, and becomes the new Leader.

## Election Timeout

During the election, the system status shows to the outside is blocking and can not respond to Client normally. Thus, the timeout should be a reasonable value such that have no effect to the normal works.

- election timeout <math>>=</math> few heartbeats
  - If the election timeout is even shorter that heartbeat, then it will be too frequent to respond to Client or sync the logs, as the logs from old Leader are more likely to be rejected.
- use random values for election timeout
  - To prevent the infinite split vote problem. It should be not that small to reduce the possible split vote times, and not that big to maintain the service. Usually 150ms ~ 300ms.

## Vote storage

Suppose that a Follower voted for itself, then down for a while, and recovered and voted for itself again, thus regardless of the votes of other Followers and became the Leader by itself.

To prevent the situation above, the vote records for Followers along with current term counts should be stored stably by themselves, and make sure that each Follower votes only once, thus ensure that there will be only 1 Leader appeared.

The machine does not need to record its previous state before voting. It will know whether it is Leader or Follower if the election continuing when it recovers, or simply relaunch an election and increase the term count otherwise.

## Log diverge

A possible log diverge situation:

![Log Diverge Example](/images/Log_Diverge.png)

Consider the (a) to (f) in the graph, assume that current Leader was down and has no possibility to recover.

(b), (e), (f) : Excluded for too small term count logs, as **larger term count server will reject the request from those smaller**.

Here each of (a), (c), (d) has the possibility to become the leader generally. And for Raft, it has own restrictions to prevent the confusion:

- The server can only vote for candidates if its last log of term count:
  - is strictly larger than local;
  - is equal to local and its log length is larger than or equal to local.


<h6 align="center">======= Lec.07 Fri. 16 Aug. 2024 =======</h6>

# Fault Tolerance: Raft (2)

- Log divergence
- Log catchup
- Persistence
- Snapshots
- Linearizability

**Leader election rule:**

- Majority
- At-least-up-to-date: **the server should have the newest term**

## Log catchup

- NextIndex: the index of the next log entry to send to the server. It is usually optimistic, which refers to that it is initialized to the length of the log of the Leader, such that the leader will think its log is the newest.
- MatchIndex: the index of the last log entry that the server has replicated. It is usually pessimistic, which refers to that it is initialized to 0, such that the server will think it has no log.

When the Leader sends the log to the server, it will send the log from NextIndex to the end of the log, and the server will compare the log with its own log from MatchIndex to the end of the log, and if the logs are the same, the server will increase the MatchIndex to the end of the log, and if not, the server will decrease the NextIndex and try again.

## Erasing log entries

The server who committed an earlier log and pulled a `log catch up` will erase the possible newer logs in other servers, and the logs will be replaced by the logs from the server.

## Log catch up quickly

- Leader sent its logs to Follower with term at `nextIndex = 6`, assume that it is 7;
- Follower received the data, found its term at `index = 5` is 5, so it sent back the reject message with its first log at term 5, assume that it is 2;
- Leader received the reject message, and decreased the `nextIndex` to 2, and sent the logs from 2 to 7 to Follower;
- Follower received the logs, and found that the logs at term 4 are the same, so it replace the later logs with the logs from Leader.

## Persistence

Consider the things happened / needed to be done when rebooting the system:

1. a server rebooted, rejoin the cluster and replay the logs.
2. a server rebooted from its persistence state(last snapshot), and catch up the logs from the Leader.

It is preferred to use the snapshot to recover the system, as it is much faster than replaying the logs. In this case, we need to know which of the state should be stored stably.

The states needed to be persisted whenever any of which changed.

**States needed to be persisted:**

- voteFor: the server can only vote once in a term.
- currentTerm: the term count of the server, which should guarantee that it is monotonic increasing.
- log: the logs of the server, to promise that the committed logs will not be withdrawn.

## Service Recovery

Similarly, there are 2 strategies to recover the service:

1. Replay the logs: too costly.
2. Periodic snapshots

Be mindful that:

- the version of the snapshots should be newer than current states.
- the newer logs should still be saved after the snapshots, and be kept when reloading from the snapshots.

## Using Raft

1. Applications integrate the packages of Raft.
2. Applications receive requests from Clients.
3. Applications call `start` of Raft.
4. Raft syncs the logs and do the operations.
5. Raft returns the results to Applications through apply channel.
6. Applications send the results to Clients.

Clients should keep a list of all the servers, and when the Leader failed, it should retry the request to another server.

Due to the possible failed requests or resending requests, the Clients should be able to handle the duplicate requests. Usually the Clients will send the request with a unique ID, and the Servers will check the ID to avoid the duplicate requests. The server used to keep these IDs is called `Clerk`.

`Clerk` is an RPC library, and it is used to ensure the list of RPC servers. If it thinks server1 is Leader, it will send the request to server1, and if server1 failed, it will retry the request to server2, and remark each request (get, put, ...) and generate a unique ID for each request to prevent the duplicate requests.

## Linearizability / Strong Consistency

1. total order of operations
   - though operations happened concurrently, the results should be the same as if they happened in a certain order.
2. match real time
   - the operations should be done in a certain time, if op1 was done before op2 started, then op1 should be put before op2 in the logs.
3. read return results of the latest write
   - if a read operation happened after a write operation, the read operation should return the result of the write operation.

<h6 align="center">======= Lec.09 Tue. 20 Aug. 2024 =======</h6>

# Zookeeper

Here requires high performance

- asynchronous
- consistency (not strong consistency)

Coordination service

**Actually here it is more likely to focus on A and P in CAP theory.**

## Zookeeper is a Replicated State Machine

Similar to Raft, Zookeeper serves as:

1. Clients visit Zookeeper, create ZNodes,
2. Zookeeper calls ZAB(similar to Raft library), generates operation logs, ZAB uses a way similar to Raft to make the cluster work, including log synchronize, heartbeats, etc...
   - ZAB:
     - maintains the order of the logs
     - prevents the network partition and split-brain problems
     - requires the operations are deterministic
3. ZAB responds to Zookeeper after finish the works
4. Zookeeper responds to the create request from Client.

ZNode is a tree structure.

As ZAB is similar to Raft, here the lecture focus on Zookeeper.

## Zookeeper throughput

If there are more write operations, then the throughput is lower,

if there are more read operations, it is higher.

The reasons for the high throughput is:

- Asynchronous
  - All operations are asynchronous, Client can submit many operations to Zookeeper at once, but the Leader will merge them and only persist once.
- Read can be done by any server
  - Every Zookeeper server can do the read operation, and no need to communicate with Leader.

As the read can be done by any server, the read operation might get the old values that has not been overwritten.

Assume a situation with 3 servers:

- A Client C1 sent `put` to make a variable `x` from `0` to `1`
- Leader L1, Follower F1 persisted the operation
- Because of the Majority, L1 sent response back to C1
- After that, another Client C2 sent `get` to query the `x`, and C2 might get :
  - `1`: C2 read from L1 or F1 with newest `x`
  - `0`: C2 read from F2 with old `x`

From the above situation, we can find that **the read / write of Zookeeper is not Linearizability**, as it is not fit with the 2nd and 3rd requirements mentioned in [Linearizability / Strong Consistency](#linearizability--strong-consistency)

Here can also make the read operation is linear though, with some additional strategies.

## Linearizability

When Client wants to connect to the Zookeeper, it will create a session, which is used to connect to the Zookeeper cluster, and its state will be maintained during the session.

1. Client creates session and connects to Leader of Zookeeper, and sends a `write` request.
2. Similar to Raft, Leader generates a log and insert it to logs, which is indexed as `zxid`.
3. Leader syncs the log to the majority number of servers like Raft.
4. Leader responds to Client, and return the `zxid` (after committed), which will be written to the Client's session.
5. Client sends a `read` request to Zookeeper, with the previous `zxid`.
6. If the requested server does not have the log of the `zxid`, it will wait until it was synced from Leader and then respond to Client.

Assume that Client sends a `write` request, and generated the `zxid=1` at Leader and Follower F1. After that, it sends a `read` request with `zxid=0` (generated by previous write operation) to Follower F2 who has the log with `zxid=0` but no `zxid=1`, it will still respond to Client with the log of `zxid=0`. However, if F2 had the log with `zxid=1`, it will send the newer log back.

Zookeeper has a `watch` mechanism. Here `watch` is like a trigger in a certain ZNode. Once the ZNode modified, Client requested the `watch` will receive an asynchronization notification, and do the correspond operations such as resending the R / W requests.

## ZNode API

Types:

- Regular: replicated everything.
- Ephemeral: temporary nodes, will be deleted when session ends or heartbeat ends, that Zookeeper thinks the ZNode has expired.
- Sequential: with version numbers and ordered by sequence id.

APIs:

- `create(path, data, flags)`
- `delete(path, version)`
- `exist(path, watch)`
- `getData(path, version)`
- `setData(path, data, version)`
- `getChildren(path, watch)`

## Summary

- Successful design
- Weaker consistency
- Suitable to be used as configuration (thanks to the APIs)
- High performance

<h6 align="center">======= Lec.10 Fri. 23 Aug. 2024 =======</h6>

# Patterns and Hints for Concurrency in Go

This part simply explained some possible implementation and optimization methods in Go, and how to improve the readability etc. of the code. Skipped.

<h6 align="center">======= Lec.11 Sat. 24 Aug. 2024 =======</h6>

# Chain Replication

## Zookeeper lock

By using the linearizablity rule of `write` of Zookeeper, we can implement a simple lock mechanism:

1. Client tries to `create` an EPHEMERAL ZNode, if success, then it gets the lock.
2. If the ZNode already exists, then the Client will wait until the ZNode is deleted by `watch` mechanism, and then try to send the `create` request again.
3. To unlock, the Client simply `delete` the ZNode.

Here although Client might crash after getting the lock, the ZNode will be deleted after the session ends (EPHEMERAL), and the lock will be released. The Zookeeper and Client will send heartbeats to each other through session, and if the Client crashed / network partition, the Zookeeper will delete the ZNode after the session ends.

However, the above mechanism is not suitable for the situation that many Clients want to get the lock at the same time, as the Clients will be blocked by the `watch` mechanism, and the unhandled requests will be sent again whenever the ZNode is deleted, which will result in a huge amount of requests.

According to the paper, there is an optimal implementation (**Ticket Lock**):

```pseudo
Lock:
1 n = create(l + "/lock-", EPHEMERAL | SEQUENTIAL) // l is the lock path
2 C = getChildren(l, false)                        // get all the children of l
3 if n is the smallest ZNode in C, exit            // if no smaller ZNode, then get the lock
4 p = ZNode in C that is just before n             // otherwise, get the ZNode just before n
5 if exists(p, true), wait for watch event         // wait for the ZNode to be deleted
6 goto 2                                           // retry

Unlock:
1 delete(n)
```

Here the `SEQUENTIAL` flag is used to ensure that the ZNodes are ordered by the sequence id, and the request will be dealt with in order, thus reduce the amount of requests.

The lock in Zookeeper is not like the lock in go. When Zookeeper ensures that the holder of the Zookeeper lock has failed it will revoke the lock, and there will be some intermediate states that the lock is not held by anyone. Here the logs of Zookeeper are used to ensure the reliability of the intermediate states.

The Zookeeper locks are usually used in:

- Master election
- Soft lock
  - Ensure that a worker of MapReduce will only have one specific Map Task at a time. As re-executing is acceptable for MapReduce, if the worker failed, the task will be reassigned to another worker.

## Approaches to build Replicated State Machines

### 1. Run all operations through Raft / Paxos (used in Lab 3)
   - This is not common in real applications, and is not the standard way to implement the Replicated State Machine as well.
### 2. Configuration manager + Primary / Backup replication (more common)
   - Implemented using a configuration manager like Zookeeper to act as Coordinator / Master, and the consensus protocol like Raft or Paxos is used inside. The Primary / Backup replication can also be used here.
   - E.g. :
     - GFS has a Master and many Chunk servers with Primary / Backup protocol
     - VM-FT has a test-and-set storage server as the Coordinator, and the VMs as the Primary / Backup, synchronized with channels or other methods.
   - It has a lower cost when maintaining the state of the system.

## Overview

This is a kind of Replicated State Machine implemented using [Approach 2](#2-configuration-manager--primary--backup-replication-more-common) above, with properties:

1. Read / query operations only involve 1 server.
2. Have simple recovery plan / mechanism.
3. The system is linearizable.

![Chain Replication](/images/Chain_Replication.png)

- The Client sends the `write` request to the Head
- Head generates the log, updates the state through storage, and sends the log to the next server.
- The Middle servers receive the log, update the state, and send the log to their next server.
- Tail receives the log, updates the state, and sends the log back to the Client.

Simply adding more servers can help to achieve a higher fault tolerance.

`write` operations are always sent to the Head, and `read` operations are always sent to the Tail for direct response.

Tails are known as `Commit Points`, as it is the only visible point to the Clients, which provides the linearizability.

## Crash

There are only 3 possible crash situations in Chain Replication:

- Head crashed
- Middle crashed
- Tail crashed

Its failover plan is easier than Raft, as the whole system is a linked list, and all the configurations are stored in the configuration server.

### Head crashed (Easiest)

If the Head crashed when the Client sent the `write` request, the Client will retry the request to the next server, it will become the new Head if valid, and the system will continue to work with the new Head.

### Middle crashed (More complex)

If the Middle crashed, the configuration server might notify the servers to build a new chain and do the extra synchronization to ensure the consistency.

### Tail crashed (Relatively easy)

If the Tail crashed, the configuration server will notify the previous server to become the new Tail, and the system will continue to work. Client can know the new Tail through the configuration server.

## Add replica

We can simply add a replica to the end of the chain, and the system will continue to work.

1. The original Tail will send the log to the new Tail and record the logs synced to the new Tail, while keeping the interaction to the Client.
2. When the sync is finished, the new Tail will send the `ack` to the original Tail through configuration manager, that it can become the new Tail.
3. Configuration manager will set the new Tail as the Tail.
4. Client can know the new Tail through the configuration manager, and the system will continue to work.

**Is it possible that the original Tail will always be the Tail, as while the new Tail is syncing, the original Tail will still interact with the Client and get the new data?**

- Assume that the original Tail is still the Tail, and the new Tail is syncing the 1 ~ 100 logs, and during which the original Tail received the 101 ~ 110 logs. Then the new Tail can inform the original Tail to stop the interaction with the Client after the sync, and inform the configuration manager that it is the new Tail, but will not be able to respond to the Client only after the sync is finished. So here might have a short block / delay for the Client.

## Chain Replication vs. Raft

+ Chain Replication
  - Split RPCs into 2 parts: `write` for Head and `read` for Tail
  - Head only needs to send the log to the next server once, and the rest of the servers will do the same.
  - `read` operations are faster as it only involves the Tail.
  - Simple failover plan.
+ Raft
  - No need for reconfiguration if any server crashed.
  - Can continue to work as long as the majority of the servers are still working.

## Extension for read parallelism

To improve the read performance, we can split object across multiple chains, and the Client can send the `read` request to any of the Tails, and the Tails will send the response back to the Client.

E.g. Here are 3 servers S1 ~ S3, and we can build 3 chains:

- S1 -> S2 -> S3
- S2 -> S3 -> S1
- S3 -> S1 -> S2

The data can be split into 3 shards, and if the shards were equally `write` to the 3 chains, the `read` operations can be parallelized to the 3 chains, and the read throughput can be improved linearly, and ideally, the read throughput can be 3 times of the original, while it can still maintain the linearizability.

Client might can know the configuration of the chains through the configuration manager, and send the `read` request to the Tails of the chains.

<h6 align="center">======= Lec.12 Sun. 25 Aug. 2024 =======</h6>

# Frangipani

- Cache coherence
- Distributed locking
- Distributed crash recovery

For traditional network file systems, the works done in servers are usually complex, while clients only need to call the APIs or cache the data.

## Overview

Frangipani has no specific role of file server, all clients themselves are acting as the file servers, as they run the file server code themselves.

Here all clients shared a single virtual disk that realized using Petal, built with many machines that replicated the disk blocks, and the consensus algorithm like Paxos is used inside to ensure the orders.

The APIs of Frangipani are read / write blocks of the virtual disks.

## Use case

It is mainly used between the researchers that temporarily need to transfer some shared files, and as all the participants are reliable, the safety of the system is not that important.

There are 2 kind of sharing:

- user to user: client to client
- user to workstation: client to server

Thus, there are some design requirements:

- caching
  - The data should not be stored all in Petal. Clients use `write back cache` instead of `write through cache`, that is, the data will be stored in the cache first, and then be written to Petal in the future.
- strong consistency
  - If one client wrote the data, the other clients should be able to see the data changes.
- performance
  - The system should be able to handle the high throughput of the data transfer.

Comparing with GFS, GFS does not provide Unix or POSIX compatibility, but Frangipani can run the Unix standard applications, and it can act as a single file system rather than a distributed file system.

## Challenges

Assume that Workstation1 does the `read f(ile)` operation, then local cache manipulate the file, there are some possible situations need to be considered how to deal with:

1. Workstation2 `cat f`: use **cache coherence** to ensure that Workstation2 can see the changes made by Workstation1.
2. Workstation1 creates `dir/f1`, Workstation2 creates `dir/f2`: need to ensure that the operations are atomic, thus files will not be replaced by each other.
3. Workstation1 crashes during file system operations: need the **crash recovery** mechanism to ensure that the system can recover from the crash.

## Cache coherence / consistency

The lock server maintains a table that records the file inode and the workstation that owns its lock. The lock server itself is a distributed service like Zookeeper, which provides APIs like `lock` and `unlock`, and has fault tolerance.

<table>
  <tr>
    <th>File inode</th>
    <th>Workstation</th>
  </tr>
  <tr>
    <td>f1</td>
    <td>WS1</td>
  </tr>
  <tr>
    <td>f2</td>
    <td>WS2</td>
  </tr>
</table>

Meanwhile, the workstations themselves need to maintain a table that records the file inode and the state of its lock.

<table>
  <tr>
    <th>File inode</th>
    <th>State</th>
  </tr>
  <tr>
    <td>f1</td>
    <td>busy</td>
  </tr>
  <tr>
    <td>f2</td>
    <td>idle</td>
  </tr>
</table>

The lock with `idle` state is called **sticky lock**, means the file has not been modified during the lock.

As the workstation owns the sticky lock, if it wants to use the file, it can directly use it without communicating with Petal or reloading cache.

Using the 2 kind of locks along with the rules below can ensure the cache coherence:

- Acquire the lock before caching the file

## Protocol

Assume that there are WS1(workstation1 / client1), LS(lock server), and WS2(workstation2 / client2).

There are 4 messages used in communications between WS and LS:

- Requesting a lock
- Granting a lock
- Revoking a lock
- Releasing a lock

The procedure as follows:

1. WS1 sends the `request` message to LS to request the lock of `f(ile)`.
2. LS receives the request, and checks the table, if `f` has not been locked, it will grant the lock to WS1, and set the owner of the lock to WS1.
3. WS1 receives the `grant` message, and caches `f`, and set the state of the lock to `busy`. After the `write` operation, it will set the state of the lock to `idle`. Here lock has an expiration time, and if WS1 crashed, the lock will be revoked by LS.
4. WS2 sends the `request` message to LS to request the lock of `f`.
5. LS receives the request, and checks the table, sends the `revoke` message to WS1. 
6. If WS1 ensures that `f` will not be modified, it will release the lock, and sync `f` to Petal, and send the `release` message to LS. Otherwise, WS2 will wait until WS1 finishes the operation and releases the lock.
7. LS receives the `release` message, records the owner of the lock as WS2, and sends the `grant` message to WS2.

The WS need to acquire the lock before accessing the file in Petal, and will sync the file to Petal after the operation, thus the cache coherence can be ensured.

## Atomicity

Assume that WS1 creates `dir/f1`, as described in paper, it will acquire the lock of `dir` and then `f1` as the pseudocode below:

```pseudo
acquire("dir")
create("f1", ...)
acquire("f1")
    allocate inode
    write inode
    update directory("f", ...)
release("f1")
```

The lock here protects the operations involving inode in the Unix file system, and ensures that the operations are atomic.

Before WS replying the LS, as both the directory and inode need the locks, it will need to firstly release the locks of `f1` and `dir`.

## Crash recovery

Updating the state in Petal needs to follow the protocol of **write-ahead logging**:

In the implementation of Petal, the disk of a machine is composed of 2 parts:

- log
- file system

When the WS wants to update the file, it will firstly write the log, and then install the update to the file system.

Every Frangipani server has its own log, and each log entry has a sequence number. The log entry stores an array of updates to describe the operations, including:

- block number that needs to be updated
- version number
- new bytes

When LS sending the `revoke` message to WS, WS will firstly send the log to Petal, then send the updated blocks to Petal, and finally release the lock.

The file data will not be written to log, but will be directly sent to Petal. The updates through logs are metadata, which is the information of the file, such as the inode, directory, etc. This is because the file data is too large to be stored in the log, which will result in low efficiency while syncing the logs.

As the file data is not stored in the log, the FS should use its own mechanism to ensure the atomicity of the file data.

For atomicity, the FS will firstly write the data to a temporary file, and then atomic rename the temporary file to the target file.

To ensure that the file operations are atomic, every log entry will have a **checksum** to ensure the integrity of the log.

## Crash scenarios

1. WS crashed before sending the log to Petal
   - The log will be lost, and the file will not be updated.
2. WS crashed after sending the log to Petal
   - The log will be stored in Petal. LS will wait until the lock expires, and ask other WSs to read the log of the crashed WS through recovery demons and apply the logged operations. After that, the lock will be reassigned to the new WS.
   - The data written by users can not be guaranteed. The log only used to ensure the metadata consistency.
3. WS crashed during writing the log to Petal
   - The checksum after the crash will be incorrect, and recovery demons will stop at the incorrect log.

## Logs and version

Assume that there are 3 WSs from WS1 to WS3:

1. WS1 logged `delete("d/f")`
2. WS2 logged `create("d/f")`
3. WS1 crashed
4. WS3 starts the recovery demon for WS1, trie to execute the `delete("d/f")` operation, but it will fail as the version of its log is smaller than the version of the log of WS2.

<h6 align="center">======= Lec.13 Thu. 29 Aug. 2024 =======</h6>

# Distributed Transaction

- 2-phase locking (2PL, suitable for a single multicore machine)
- 2-phase commit (2PC, suitable for distributed systems)

The reason for using distributed transaction is that human need cross-machine operations, and the operations need to be atomic.

## Transactions

```markdown
T1:             T2:
- begin(x)      - begin(x)
| - add(x, 1)   | - get(x)
| - dec(y, 1)   | - get(y)
- commit(x)     | - print(x, y)
                - commit(x)
```

A transaction is a sequence of operations that are **Atomic, Consistent, Isolated, and Durable** (ACID). See [Wikipedia-ACID](https://en.wikipedia.org/wiki/ACID) for more details if not familiar.

## Concurrency control

- Pessimistic: Lock before read / write to ensure the serializability.
- Optimistic: Read / write without lock, and check if the operations are serializable when reaching the commit point. If not, rollback / abort the transaction and retry(need other mechanisms).

Here focuses on the pessimistic concurrency control.

## 2-phase locking (2PL)

**Lock per record**

- T acquires the lock before using
- T holds until commit / abort 
- **The number of locks will only increase during the transactions using 2PL**, this is to ensure that the intermediate states are consistent and not visible to other transactions.

2PL is an improvement of simple(strict) locking. A simple locking or strict locking will lock all the records at the beginning of the transaction, and release all the locks at the end of the transaction. Instead, 2PL will dynamically acquire the locks when needed, which supports a more flexible concurrency. Generally, 2PL has a higher concurrency than simple locking.

### Deadlock

Consider a situation that:

| time | T1                 | T2                 |
|------|--------------------|--------------------|
| 1    | lock(x), put(x)    |                    |
| 2    |                    | lock(y), get(y)    |
| 3    |                    | wait for x, get(x) |
| 4    | wait for y, put(y) |                    |

Here T1 and T2 are in a deadlock that T1 is waiting for lock(y) and T2 is waiting for lock(x).

To solve the deadlock when detected, the transaction system will choose one of the transactions (victim) to abort. The victim will roll back and release the locks, and the other transaction will continue to work. The client or application will decide how to deal with the aborted transaction, like retrying or simply giving up.

There are 2 ways to detect the deadlock:

- Timeout: If the transaction waits for a long time, it will be considered as a deadlock.
- Wait-for graph: The system will build a graph to record the waiting relationship of the transactions, and if there is a cycle in the graph, it will be considered as a deadlock.

## 2-phase commit (2PC)

<table>
  <tr>
    <th>Time</th>
    <th>Coordinator</th>
    <th>Server A (with X record)</th>
    <th>Server B (with Y record)</th>
  </tr>
  <tr>
    <td>1</td>
    <td>Create Transaction<br>inform A to execute put(x)<br>inform B to execute put(y)</td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>2</td>
    <td></td>
    <td>Lock x<br>put(x)<br>log(x)</td>
    <td>Lock y<br>put(y)<br>log(y)</td>
  </tr>
  <tr>
    <td>3</td>
    <td>Query:<br>?prepare(A)<br>?prepare(B)</td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>4</td>
    <td></td>
    <td>prepare(A)</td>
    <td>prepare(B)</td>
  </tr>
  <tr>
    <td colspan="4">Case: <code>prepare(A) && prepare(B)</code></td>
  </tr>
  <tr>
    <td>5</td>
    <td>Commit:<br>commit(A)<br>commit(B)</td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>6</td>
    <td></td>
    <td>Install log (execute put(x))<br>Unlock x<br>OK</td>
    <td>Install log (execute put(y))<br>Unlock y<br>OK</td>
  </tr>
  <tr>
    <td>7</td>
    <td>Acknowledge OKs</td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td colspan="4">Case: <code>!prepare(A) || !prepare(B)</code> (due to deadlock, lack of space, etc.)</td>
  </tr>
  <tr>
    <td>5</td>
    <td>Abort:<br>abort(A)<br>abort(B)</td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>6</td>
    <td></td>
    <td>Rollback(x)<br>Unlock x<br>OK</td>
    <td>Rollback(y)<br>Unlock y<br>OK</td>
  </tr>
  <tr>
    <td>7</td>
    <td>Acknowledge OKs</td>
    <td></td>
    <td></td>
  </tr>
</table>

## 2PC crash analysis

- Server crashed after replying `?prepare: OK`
  - The crashed server will continue to work after recovery and check logs to load the data. 2PC in this case will have some expensive procedures to ensure the consistency, like sending multiple messages and stably storing operation logs.
- Coordinator crashed after sending `commit`
  - If A replied `OK`, but B did not, then B will wait for the coordinator to recover and send the `OK` message again. In this case, B has to wait, during which the `lock(y)` will be held.
- Server did not reply the `?prepare` message
  - The coordinator will wait. If timeout, the coordinator will abort the transaction, and will inform the server to terminate the transaction when the server recovers.

To enhance the fault tolerance, the Raft or other consensus algorithms can be used in the 2PC, and the logs can be stored in the Raft to ensure the C(consistency).

Raft is actually similar to 2PC. The biggest difference is that Raft based on the rule of majority, while 2PC requires all the servers to have same replies. Also, for Raft, all the servers are doing the same thing, while in 2PC, the servers are doing different things. Meanwhile, Raft is for high availability, while 2PC is for atomicity across machines.

<h6 align="center">======= Lec.14 Fri. 30 Aug. 2024 =======</h6>

# Spanner

- Wide-area transactions
  - Read-write transaction 2PC+2PL+Paxos groups
  - Read-only transaction: Optimized by Spanner, about 10x faster than read-write transaction, can run in any data center.
    - Snapshot isolation
    - Synchronized clocks (TrueTime)
- Wide-used in Google

## Organization

Suppose that there are 3 servers S1 ~ S3, which have same shards that belong to the same Paxos group:

| Paxos group | S1          | S2          | S3          |
|-------------|-------------|-------------|-------------|
| Group #1    | Shard a ~ m | Shard a ~ m | Shard a ~ m |
| Group #2    | Shard n ~ z | Shard n ~ z | Shard n ~ z |

Here multiple shards is for higher parallelism. If transactions involved multiple shards, a transactions will not depend on other transactions, and can be executed in parallel.

Every shard has its own Paxos group that used for replication, and can run normally with the majority of the servers. The replication is for the fault tolerance, and using majority rule can pass through slowness.

Replicas were usually deployed at somewhere close to clients, such that client can access its nearest replica. The read-only transactions are usually executed by local replicas without communication with other data centers.

## Challenges

- Read of local replica yield latest write, requires a stronger consistency.
- Support transactions across shards.
- Read-only transactions and read-write transactions must be serializable.

## Read-write transactions

- When executing a read-only transaction, the client will directly read the leader of the Paxos group.
- When executing a read-write transaction, the client will need to do write operations through the coordinator that coordinates the leaders between Paxos groups.
- Lock table only stored in the leader of Paxos group, and does no replication in the group(to accelerate the read-only transactions).
  - If leader crashed, the transaction will need to restart as the lock table is lost.
  - If crashed after `prepare`, the state of some records with locks will be synced to other group members, and will not be lost.
- The coordinator of 2PC is also the Paxos group, this is for enhancing the fault tolerance and reducing the possibility that the coordinator crashed and participants have to wait to commit.

Multiple servers as Coordinator, multiple servers execute the operations related to shards.

| Time | Client                                                                | Paxos-Coordinator                          | Paxos-Shard-A(Sa)                                                | Paxos-Shard-B(Sb)                                                |
|------|-----------------------------------------------------------------------|--------------------------------------------|------------------------------------------------------------------|------------------------------------------------------------------|
| 1    | Transaction ID: tid<br>To Sa Leader: read(a)<br>To Sb Leader: read(b) |                                            |                                                                  |                                                                  |
| 2    |                                                                       |                                            | Leader lock(x) by 2PL<br>Set owner as Client                     | Leader lock(y) by 2PL<br>Set owner as Client                     |
| 3    | Send to Coordinator:<br>`add(x), dec(y)`                              |                                            |                                                                  |                                                                  |
| 4    |                                                                       | Received the request<br>`?add(x), ?dec(y)` |                                                                  |                                                                  |
| 5    |                                                                       |                                            | R-Lock => W-Lock<br>`log(add(x))`                                | R-Lock => W-Lock<br>`log(dec(y))`                                |
| 6    |                                                                       | `?prepare(Sa)`<br>`?prepare(Sb)`<br>`log`  |                                                                  |                                                                  |
| 7    |                                                                       |                                            | `OK` if: <br>locked, <br>logged, <br>synced, <br>ready to commit | `OK` if: <br>locked, <br>logged, <br>synced, <br>ready to commit |
| 8    |                                                                       | `commit(tid)`<br>`log`                     |                                                                  |                                                                  |
| 9    |                                                                       |                                            | install log, <br>execute `add(x)`, <br>unlock x, `OK`            | install log, <br>execute `dec(y)`, <br>unlock y, `OK`            |

Shards replicating the lock that is holding when it does the `prepare` operation rather than directly replicating the whole lock table.

## Read-only transactions

- Fast, read from local shards
- No need to lock, no need to do 2PC

### Correctness

Here correctness means:

1) the transactions are serializable
2) external consistency
   - if a transaction T2 starts after T1 commits, then T2 should see the results of T1
   - **serializability + real-time ordering**, more like the linearizability

The difference between external consistency and linearizability is that linearizability is operation-based, while external consistency is transaction-based.

### Bad plan

| Time | Transaction T1       | Transaction T2       | Transaction T3           |
|------|----------------------|----------------------|--------------------------|
| 1    | set X=1, Y=1, Commit |                      | Read X (get X=1)         |
| 2    |                      | set X=2, Y=2, Commit |                          |
| 3    |                      |                      | Read Y (get Y=2), Commit |

Here T3 should see the same X and Y, but it sees X=1 and Y=2, which is inconsistent.

To solve this, Spanner uses **Snapshot Isolation**.

### Snapshot Isolation

- Assign a timestamp to every transaction
  - Read-write transaction: the timestamp is the commit time
  - Read-only transaction: the timestamp is the start time
- Execute transactions in timestamp order
- Each replica stores data with timestamp

| Time | Transaction T1<br>(Timestamp: 1) | Transaction T2<br>(Timestamp: 3) | Transaction T3<br>(Timestamp: 2) |
|------|----------------------------------|----------------------------------|----------------------------------|
| 1    | set X=1, Y=1, Commit             |                                  | Read X (get X=1)                 |
| 2    |                                  | set X=2, Y=2, Commit             |                                  |
| 3    |                                  |                                  | Read Y (get Y=1), Commit         |

Here T3 will get the latest committed value before its timestamp, so it will see X=1 and Y=1.

Every replica maintains a version table that stores the versions of the data along with the timestamps, and the read-only transactions will read the data with the timestamp that is less than its own timestamp.

The Spanner here also depends on the **SafeTime** mechanism, which ensures that the read operation will get the latest data before its timestamp.

- Paxos sends write operations in timestamp order.
- If a read operation with timestamp Tr is sent to a Paxos group, the Paxos group will wait until any write operation committed with timestamp Tw > Tr is sent to the Paxos group, and then execute the read operation.
  - Here also need to wait for the prepared but not committed operations in 2PC.

## Clock drift

- Matters only for read-only transactions

If the timestamp is:
- too large:
  - Assume T1 executed at 15, but the machine thinks it should be executed at 18, then it will only need to wait for longer time, and the correctness will not be affected.
- too small:
  - Assume T1 executed at 15, but the machine thinks it should be executed at 9, and if there is a committed transaction T2 with timestamp 10, then T1 will not see the data of T2, which will affect the correctness.

## Clock synchronization

1. Atomic clocks
2. Synchronize with global time (using GPS or etc.)

Though applied the clock synchronization, the clock drift still exists, and Spanner uses the **TrueTime** mechanism to ensure the correctness (AbsoluteTime + (earliest, latest) uncertainty).

[NTP](https://en.wikipedia.org/wiki/Network_Time_Protocol) might be used here to synchronize the clocks.

To solve the clock drift, Spanner uses the **time interval** rather than real timestamp. Every transaction has a `now` with `[earliest, latest]`.

- Start rule: now.latest
  - read-write transaction: assign `now.latest` when commit starts
  - read-only transaction: assign `now.latest` when transaction starts
- Commit wait rule: delay until `now.earliest`

| Transaction T1(@1)             | Transaction T2(@10)             | Transaction T3(@12)              |
|--------------------------------|---------------------------------|----------------------------------|
| set X=1, Commit<br>`[@1, @10]` |                                 |                                  |
|                                | set X=2, Commit<br>`[@11, @20]` |                                  |
|                                |                                 | Read X (get X=2)<br>`[@10, @12]` |

## Summary

- Read-write transactions: 2PC + 2PL + Paxos groups, global serializability
- Read-only transactions: Read only local replica, use snapshot isolation to ensure the correctness
  - Snapshot isolation: ensure the serializability.
  - Ordered by timestamp: ensure the external consistency.
    - Using time interval.

<h6 align="center">======= Lec.15 Sat. 31 Aug. 2024 =======</h6>

# FaRM (Fast Remote Memory)

FaRM provides strict serializability, and is designed for the high-performance distributed system. (using 90 servers to achieve 140 Million TPS, comparing with 10 ~ 100 TPS of Spanner)

- One data center
- Shard
- Non-volatile DRAM
- Kernel bypass + RDMA
- OOC (Optimistic Concurrency Control)

## Setup

In a data center, there are 90 servers connected with a high-speed network, and each server has different shards called regions, each region is 2GB on DRAM (can be larger if needed), DRAMs are deployed on the servers with UPS (Uninterruptible Power Supply) to ensure the data will not be lost. They have a Primary-Backup replication mechanism to ensure the fault tolerance, with configuration manager and Zookeeper maintaining the mapping from the region to the servers (region# -> Si).

We might imagine that region is like an object array, with 2GB per object. The object has a unique ID (`oid`), contains the region# and the offset in the region, and the metadata of the object (with a 64 bits number, top 1 bit for the lock, and the rest for the version number).

## FaRM API

- `txBegin()`: start a transaction
- `txCommit()`: commit the transaction
- `read(oid)`: read the object
- `write(oid, data)`: write the object

For abort, due to the optimistic concurrency control, it will retry the transaction.

FaRM will use protocol like 2PC to apply atomic operations on objects across regions.

## Kernel bypass

FaRM runs as a user application on the operating system. It uses SSD for temporary storage when recovering from the crash, and uses DRAM to ensure the efficiency of read / write operations. To solve the bottleneck of the CPU, FaRM uses the kernel bypass mechanism to deal with the network data. Regularly this is done by using kernel drivers to read / write the NIC (Network Interface Card), and it is costly as the interaction of application and NIC will need to call the kernel and involve the network stack. The kernel bypass mechanism will order a send / receive operation queue that will be directly mapped to the address space of the application, and the application can directly read / write the NIC without calling the kernel, which will reduce the cost of the CPU. For the implementation, FaRM uses a user-level thread to poll the receiving queue of the NIC for the available data. FaRM keeps shifting between the application thread and the NIC polling thread.

## RDMA (Remote Direct Memory Access)

RDMA requires the support of NIC.

Assume 2 FaRM servers connected with cable, server1 can directly put an RDMA packet to the sending queue of the NIC, and the NIC will send the packet to the NIC of server2. Server2 will parse the commands along with the packet that describes the operations (like read / write), and send the response back to server1. 

RDMA can be used to do operations directly to the memory of the server without the help of interruption or other mechanisms, the NIC has the firmware to deal with the operations, load the data in the memory, request the memory address and put it to the response packet, and send it back. This is **one-sided RDMA**, which usually refers to the read operations.

The **write RDMA** is mostly the same as the read, but the sender can put the RDMA packet and inform that it is a write operation.

Write RDMA usage:

- Log: including the commitment of the transaction, the leader can use the write RDMA to append the log to the followers.
- RPC message queue: one queue per pair. The sender can use the write RDMA to put the message to the queue, and the receiver uses a polling thread to read the message from the queue, and respond to the sender using the write RDMA. This is more efficient than the standard RPC.

In FaRM, the one-sided RDMA spends about 5 microseconds, which is like the time of accessing local memory.

## Transaction using RDMA

As RDMA does not provide the ability to run code on servers, it needs other protocols or mechanisms to implement 2PC and transactions, canceling or reducing the usage the participation of the servers.

## OOC (Optimistic Concurrency Control)

- read objects without locking (version number)
- add validation step before commit
  - if version number conflicts, abort the transaction
- commit the transaction

If the transaction is aborted, the client will retry the transaction. The read operation can be done simply by RDMA, with no need to modify any state of the server.

![FaRM_Transaction](/images/FaRM_Transaction.png)

## Strict serializability

The strict serializability here is guaranteed by the version number.

Assume that there are 2 transactions T1 and T2, and both of them did:

```markdown
TxBegin
  o = read(oid)
  o.data = o.data + 1
  write(oid, o)
TxCommit
```

The result of `o.data` can be 0, 1, or 2 with following situations:

- `0`. one detected the conflict and aborted, while the other failed.
- `1`. one conflict or failed, while the other succeeded.
- `2`. both succeeded.

If T1 and T2 read in the same time, and T1 get the lock first:

- if T1 did not commit, T2 will abort.
- if T1 committed and T2 get the lock, when validating, T2 will find the version number is not the same as the one it read, and will abort.

Thus, if T1 and T2 did not fail, after one of them committed, the other will be aborted.

To test it, let T1 and T2 be:

```markdown
Assume x = 0, y = 0
T1:             T2:
if x == 0:      if y == 0:
    y = 1           x = 1
```

In this test case, the result can be either x = 1, y = 0 or x = 0, y = 1, but not x = 1, y = 1.

If T1 and T2 are executed in parallel:

| Time | T1                   | T2                                                                               |
|------|----------------------|----------------------------------------------------------------------------------|
| 1    | read(x), read(y)     | read(y), read(x)                                                                 |
| 2    | lock(y), validate(x) |                                                                                  |
| 3    |                      | lock(x), validate(y)<br>the top bit of the 64 bit metadata is 1(locked)<br>abort |

So it is strict serializability.

## Summary

- Fast
- Assume few conflicts
- Data should fit in the memory
- Replication is only in the same data center
- Require stable hardware

<h6 align="center">======= Lec.16 Sun. 1 Sep. 2024 =======</h6>

# Spark

- "Successor" of Hadoop, with the same MapReduce model, but with more features and optimizations.
- Widely used in the industry. (Databricks, Apache Spark)
- Wide-range of applications.
- In-memory computations.

## Programming model: RDD

- Resilient Distributed Dataset

```Scala
// 1.create an RDD: lines. Here RDD is stored in HDFS.
lines = sparks.textFile("hdfs://...")
// 2.create another RDD: errors, lines are read-only, and errors are filtered and stored to the new RDD.
errors = lines.filter(_.startsWith("ERROR"))
// 3.persist the RDD(errors) in memory for future use
errors.persist() // Different from MapReduce, which will write and read the intermediate results in intermediate files.

// 4. Count errors mentioning MySQL. This is an action, and will trigger the computation of the RDD.
errors.filter(_.contains("MySQL")).count()

// 5. Return the time fields or errors mentioning HDFS as an array
// assuming time is the first field in a tab-separated format
errors.filter(_.contains("HDFS"))
      .map(_.split("\t")(0))
      .collect() // collect is another action that will trigger the computation of the RDD
```
`lines (filter)=> errors (filter)=> HDFS (map)=> time (collect)=> Array`

The APIs of RDD support 2 kinds of Operations:

- Transformation: transform the RDD to another RDD. **Every RDD is immutable(read-only)**, and the transformation can only generate a new RDD from the existing RDD.
- Action: trigger the computation of the RDD. Only after executing the action, the Transformation will be executed. If using the command line, the result will be printed only after the action is executed.

![RDD_APIs](/images/RDD_Transformations_Actions.png)

## Fault tolerance

- Narrow dependency: parentRDD and childRDD is one-to-one
- Wide dependency: parentRDD and childRDD is one-to-N


### Narrow dependency

This is mostly the same as MapReduce. If a worker of Spark failed, the stage will be re-executed, and the partition of the RDD will be reread or recomputed.

### Wide dependency

Assume that a Spark is in a command flow, `childRDD1` has data from `parentRDD1 ~ parentRDD3`, and operations like `join` are done and generate `childRDD2`. If the worker generating `childRDD2` failed, according to the Narrow dependency fault tolerance, the whole process will be re-executed, and the data of `parentRDD1 ~ parentRDD3` will be re-read or re-computed. This will be inefficient and costly.

The solution here is to set checkpoints or persist the intermediate results of the RDDs.

## Iterative: PageRank

PageRank is an iterative algorithm that calculates the weight / importance of the web pages. The weight of a page is the sum of the weights of the pages that link to it.

```Scala
val links = spark.textFile(...).map(...).persist()
var ranks = // RDD of (URL, rank) pairs
for (i <- 1 to ITERATIONS) {
    // build an RDD of (targetURL, float) pairs
    // with the contributions sent by each page
    val contributions = links.join(ranks).flatMap {
        (URL, (links, rank)) => 
          links.map(dest => (dest, rank / links.size))
    }
    // sum contributions by URL and get new ranks
    ranks = contributions.reduceByKey((x, y) => x + y).mapValues(v => a / N + (1 - a) * sum) // if add .collect(), it will trigger the computation
}
```

`links` stores the connections of the graphs (URL pointing relationships), and `ranks` stores the rank of the URLs.

![PageRank](/images/PageRank.png)

In the image, the links join the ranks and generate a new stage of RDD. This can be optimized to parallel by partitioning the RDDs using hash partitioning, thus the `join` operation can be done like narrow dependency.

The only need to persist the `links`, as things like `ranks1`, `ranks2`, etc. are new RDDs. For certain situations, the `ranks` can be persisted to avoid re-computation if crashed.

The `contribs` can be calculated in parallel if partitioned, and they will finally be reduced by the `.collect()` operation.

## Summary

- RDD: Resilient Distributed Dataset, made by functional transformations.
- group together in sort of lineage graph
  - allows reusing the intermediate results
  - clever optimizations by coordinator
- more expressive than MapReduce
- in memory for better read / write performance

Each worker runs a stage in a partition. All stages run in parallel on different workers. Every stage in pipeline is a batch of tasks.

<h6 align="center">======= Lec.17 Mon. 2 Sep. 2024 =======</h6>

# Memcached

From the paper by Facebook in 2013, we can learn:
- Impressive performance
- Tension between consistency and performance (the application of Facebook does not need real linearized consistency)
- Cautionary tale

## Website evolution

```markdown
   single   server  + single   database 
-> multiple servers + single   database 
-> multiple servers + multiple databases (sharding, db parallelism, 2PL + 2PC)
-> multiple servers + multiple databases + cache (cache layer)
``` 

For `multi-servers + multi-databases + cache`, every single cache server is called **Memcached daemon**, and the whole cache cluster is called **Memcached**.

The main challenges here are:

1. how to maintain the consistency between the cache and the database
2. how to guarantee that the database will not be overloaded

## Eventual consistency

- write ordering (by database)
- reads are behind is fine
  - clients should be able to read their own latest writes

This is similar to Zookeeper, where the different sessions can possibly not able to read the latest data written by other sessions, but can read their own latest writes.

## Cache invalidation

![Memcached](/images/Memcached.png)

1. Front-end writes the data to MySQL
2. Squeal daemon monitoring the transaction log of MySQL, and finds a key that is modified
3. Squeal daemon sends the message that the key is invalid to the Memcached
4. Memcached deletes the cached data of the key
5. Front-end reads the data from Memcached
6. If the data is not found in Memcached, it will read from MySQL, and write the data to Memcached

## Fault tolerance

Facebook deployed 2 data centers including the primary data center and backup data center.

- the Primary supports the read and write operations, and the Backup supports the read operations only
- if the Backup receives the write operation, it will request the Primary.
- the changes in the Primary will be monitored by the Squeal, and the changes will be synced to the Backup.
  - the short latency exists, but acceptable (not strong consistency)

## High performance

- Partitioning or sharding
  - Capacity: Every shard can store a large amount of data
  - Parallelism: Different shards can be accessed in parallel
- Replication
  - hot key: Same key can be extended to different Memcached servers to ease the pressure of hot key access
  - Capacity: Need to enlarge the capacity of the Memcached servers for replication
- Cluster
  - Popular key: The popular key can be stored in the Memcached servers that are in the same cluster, and the key can be accessed in parallel
  - Reduce connection: Prevented in-cast congestion by averagely distributing the TCP connections from the front end to the Memcached servers
  - Reduce network pressure: It is hard to build bisection bandwidth networks that can sustain a huge load

There is a regional pool that is used to store the keys that are not popular or infrequently used.

## Protecting DB

If all Memcached servers are down, the database might be overloaded when receiving too many requests.

### New cluster
  - If a new cluster that is supposed to handle 50\% read requests is built but has not been warmed up, the requests will be sent to the database, which will cause the database to be overloaded. Thus, the new cluster should be warmed up before being used.
### Thundering herd
  - If a hot key is updated and its cache is invalidated, all the requests will be sent to the database, which will cause the database to be overloaded. Here facebook uses **lease**, that when a hot key is invalid, a front end will get a lease and be responsible for updating the cache, and the other front ends will retry the request to the cache rather than the database.
### Memcached server failure
  - Set a **gutter pool** that is used to temporarily handle the situation when a Memcached server is down, and the requests will be first sent to the gutter pool, and then be sent to the database if the data is not found in the gutter pool.

## Race

### Stale set

| Time | Client 1                | Client 2                        |
|------|-------------------------|---------------------------------|
| 1    | get(key)<br>get lease   |                                 |
| 2    |                         | delete(key)<br>invalidate lease |
| 3    | put(key, v)<br>Rejected |                                 |

Here C2 updated the database by deleting the key, so the lease will be invalidated, and C1 will fail to update the cache. This prevents the stale set and the [Thundering herd](#thundering-herd).

### Cold cluster

| Time | Client 1                            | Client 2                                                                              |
|------|-------------------------------------|---------------------------------------------------------------------------------------|
| 1    | delete(key)<br>*two-second hold-off |                                                                                       |
| 2    |                                     | get from cold cluster(not found)<br>get from warm cluster<br>put(key, v)<br>permanent |

*Two-second hold-off: If the key is deleted from a cold cluster, the put(key) will not be executed in the next 2 seconds, so the `put(key, v)` might be rejected. This will only happen when the cluster is warming up, and 2 seconds is enough for synchronization.

### Regions (Primary and Backup)

C1 writes to the Backup, and the data will be written to the Primary, then execute `delete(key)` to invalidate the cache in the Backup. At current stage if C1 execute `get(key)`, it will get `null` as the data requires some time to be synced to the Backup by Squeal.

Facebook uses **remote marker**, that when `delete(key)` is executed, it will mark key as "remote", and the `get(key)` will find the mark and read the data from the Primary. After finishing the `get(key)`, the "remote" mark will be removed.

## Summary

- Caching is vital
- 
  - Partitioning and sharding
  - Replication
- Consistency between memcache and database, `lease` / `two-second hold-off` / `remote marker`

<h6 align="center">======= Lec.18 Tue. 3 Sep. 2024 =======</h6>

# Fork Consistency - SUNDR (Secure UNtrusted Data Repository)

**Decentralized systems**

For such systems, the designer needs to consider or account for the **[Byzantine fault](https://en.wikipedia.org/wiki/Byzantine_fault)**, which is a fault that might occur in a distributed system that might cause the system to behave in an unexpected way.

SUNDR proposes some powerful ideas like signed logs that can be used in the decentralized systems.

## Setting: Network file system

SUNDR is designed for the network file system, where the clients can read and write the files. This is similar to the Petal in Frangipani, and its procedure of the file operations are mostly the same as Frangipani. The difference is that SUNDR supposed that the network file system can be byzantine, and the potential attackers might be able to master the NFS using some techniques. For this reason, the clients in SUNDR are not trusted, neither the file system.

Possible attacks:

- bugs in software
- system administrator with weak password
- physical break-in
- bribes operators / colludes with malicious clients

## Focus: Integrity

Like Debian Linux was once being attacked and its code was injected with trapdoors. And the development was frozen in order to make sure the integrity of the code, that which part of the code is right and which part has been modified.

The integrity is to make sure that the system structure is right, and the data is not modified by the attackers.

## Safety problem example

Assume zoobar is a virtual bank application, and the users can transfer the zoobars to each other as a kind of currency.

auth.py: the authentication module
bank.py: the bank module

```markdown
A: someone responsible for the authentication, and has the permission to modify the auth.py
B: someone responsible for the bank, and has the permission to modify the bank.py
C: someone copied the project, deployed it, and was unfortunately attacked by the attackers
```

Consider the following situation:

```markdown
S modified the auth.py, such that the project deployed by C will lack some authentication steps
S modified the bank.py to delete the authentication steps
```

The common solutions to prevent the above situation are as follows:

Use asymmetric encryption to sign the code (public key and private key), and that S cannot modify the `auth.py` as S does not have the private key of A, and the signature of the code will be verified and eventually be rejected by the code repository.

However, here S still has some methods to attack the system, such as sending the previous version of `auth.py` as it is signed by A and might have some possible bugs that make it vulnerable. Also, S can declare that the `auth.py` has been deleted, and C will not be able to verify if it is true. This might need to pack up all the files, and decide the newest version of the files.

## Big idea: Signed logs of operations

Assume that a log like:

| mod<br>auth.py<br>signed by A | mod<br>bank.py<br>signed by B | fetch<br>auth.py<br>signed by C | fetch<br>bank.py<br>signed by C |
|-------------------------------|-------------------------------|---------------------------------|---------------------------------|

By using the logs, the system can verify the operations with the operators, and if someone received a log, it will not able to lose the previous logs by comparing the signatures. It is also harder for the attackers to delete some logs.

The procedure of a user downloading the project:

1. checks signatures
2. check its own last log entry
3. construct FS
4. add the new log entry, and sign it
5. upload the new log entry to FS

**Why fetches in the log?**

If the fetch is not in the log, assume that `auth.py` and `bank.py`, if C fetches them with different versions, there might have some problems when running.

While if logged:

1. C fetches `auth.py`, received the log entry of A and B
2. C adds the log, uploads the `fetch` log entry to FS
3. C fetches `bank.py`, received the log entry of A and B, and the `fetch` log entry of C. If the `fetch` log entry is not found, the system will reject the fetch operation.

## Fork consistency

The server can not really operate the logs, it can only send prefixes or hide parts of the logs. So when logs are modified, different clients might receive different logs.

Assume that the logs are like:

| Server S | a |   |   |
|----------|---|---|---|
| Client A | a | b | c |
| Client B | a | d | e |

Here clients have same log prefix, and S will keep both logs rather than merge them.

## Detecting forks

1. out-of-band communication: the clients can communicate with each other to check if the logs of one are the prefix of the other
2. timestamp box: introduce a timestamp box that adds the timestamp to the logs every several minutes (similar to fork), and the clients can check the timestamp to see if the logs are forked

Bitcoin: settle forks, like the 2nd above, using some methods to achieve the consensus and decide which fork can be used.

## Snapshot per user and version vector

![User_Snapshot](/images/User_and_Group_i-handles.png)

SUNDR uses the snapshot per user, and the version vector to handle the fork consistency problem.

The server maintains logs, and others maintain the snapshots.

An i-handle is used to represent the snapshot of the user, it maps to an i-table hash table that stores all inodes and their hash values. If a file was modified, the hash value of its corresponding inode will be updated, and the i-table, i-handle will be updated as well, which serves a complete snapshot.

The version vector is used to handle the consistency between users. Every version vector maps to an i-handle, and maintains the modify count of all users.

```pseudo
VS_A: i-handle-A {A: 1, B: 0, C: 0}
VS_B: i-handle-B {A: 1, B: 1, C: 0}

C: By comparing the version vectors, C will choose the newest, here B, and fetch its file system snapshot.
```

## Summary

- Byzantine participants
- Signed logs
