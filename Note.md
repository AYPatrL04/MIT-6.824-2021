<h1 align="center">MIT 6.824 Learning Notes</h1>
<h5 align="center">AYPatrL04</h5>
<h5 align="center">AYPatrL04@gmail.com</h5>

<h6 align="center">======= Lec.01 Fri. 02 Aug. 2024 =======</h6>

# Introduction

## Labs
1) MapReduce
2) Replication Using Raft
3) Replicated Key-Value Service
4) Sharded Key-Value Service

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
  - mapreduce deals with distribution

# Abstract view
**Mapreduce**: an idea come from functional programming.
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
4. The `Map` workers read the input files, and turn them into a list of key/value pairs (store intermediate key/value files on local disks).
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
  - I/O concurrency
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
- the procedure of marshal and unmarshal parameters happens in the stub.
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

Usually the replication is used to tolerate failures. However the amateur replication might causing the readers to read the different versions of the data.

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

#### Works:

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

### If append operation failed, and the client retries the operation, will the offset be the same with the previous?

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

### How to get stronger consistency?

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

Common failure scenaries:

- Fail-Stop failures: infrastructure failure or components of the computer does not work well and eventually stops the computer.
- Logic bugs, configuration errors
- Malicious errors
- Handling stop failure

Here we assume that there are no errors with the system and softwares, no logical errors or configuration errors, and no attackers. In other word, we only focus on **Handling stop failure**.

## Challenge

- Has primary failed?
  - the machine might fail to visite the primary due to network partition or machine failed, and in this case some strategies should be applied in adavance to avoid the co-existence of 2 primaries (assume that there are only 1 primary allowed).
- Split brain system
  - if each partition of networks has its own primary, and clients interact with these primaries, there might be huge difference in data, version, etc., and if the system restart, we have to handle it by ourselves (similar to the handling of git merge).
- Keep primary/backup in sync
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

Most of systems choose the 2nd approach, as it would be more expensive to sync the possible gigabytes of states generated by the operations. It would be much less expensive if backup can do the operations and generate the states itself.

GFS also applied the 2nd approach (RSM). See [GFS File Write Procedure](#gfs-file-write-procedure) Step 4 - 7.

## Level of operations to replicate (RSM)

- Application-level operations 
  - Such as GFS file append or write. If using RSM on aplication-level operations, the semantics of these operations should also be known by RSM, and it should know how to deal with these operations. <ins>if applying a RSM approach on the kind of application-level, <strong>the Application Itself</strong> should be modified to perform or play part of the RSM approach</ins>
- Machine-level operations
  - The state here refers to the state of registers or memories, and the operations are traditional computer instructions. In this level, the RSM only focus on the most basic machine instructions rather than applications or operation systems.
  - A traditional way of applying machine-level operations is to buy those machines or processors that support replicate/backup itselves, but it is usually expansive.
  - Instead of hardware replication, Virtual Machine can be used here.

## VM-FT: Exploit virtualization

The virtual replication is transparent to applications, and appears to clients as if the system is a single machine **(Strong consistency)**. VMWare is based on this in early versions. The shortcoming is that this supports only single-core rather than multi-core. Maybe it is supported by later VM-FT, but the solution might be State Transfer rather that pure RSM.


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
2. send the interruction signal to the actual VM, like the Linux VM running on guest space as shown in the graph.

Similarly, when client send packets to primary and the hardware of primary occurred interruptions, VM-FT will send the interruption to backup through logging channel, and send it to the current VM. The VM then writes the data to the virtual network card in VM-FT, and the VM-FT will write the data to the physical network card to send off the response to the client. Meanwhile, the backup also receives the interrption send by the primary VM-FT, but it will do nothing when its virtual network card received the data.

Assume that there is a storage server outside and connected to the VM-FTs through network, the storage use a flag to record the state of primary and backup, and who is primary.

When primary and backup occurs the network partition problem while they are still able to communicate with the storage server, they will assume that the opposite machine is down, and trying to act as the new primary to change the flag in storage through the test-and-set (sets a new value and returns the original value) atomic operation, and whoever finished the operation in advance will be marked as new primary. The latter one will receive the changed value and realize that it is the latter one, and give up becoming primary (terminate itself).

- When is the flag be initialed?
  - Firstly a primary is started, and a backup is started for backup. Then a repair plan is needed to guarantee that if primary fails, the backup can take its place. In VM-FT, the repair plan is executed manually by monitor software. It creates a new replica based on the VM image, and when the backup started and finished the backup, the flag is reset according to the protocol. 
- What will happen if the logging channel or the channel the client used to access the server breaks?
  - If the channel broke, the server will no longer serve the client as long as primary and backup themselves work well, and nothing can be done except wait until the network being repaired.
- A kind of logging channels were implemented by UDP for efficiency. If it fails like a certain packet was not confirmed, will primary simply think backup fails and performs no redial?
  - No because of the timer interruptions. It triggers like every 10ms, and if there is a packet failed to be received, primary will try to resend it to backup. If it is still occurring problems, the work might be stopped.

## Divergence sources

- Non-deterministic instructions
  - E.g. the instruction to get the time as it is hard to assure that the primary and backup execute in the same time, and the received values are usually different
- Input packets / timer interruptions
  - E.g. the interrupters might be inserted into the instructions stream between 1st and 2nd instructions of primary while between 2nd and 3rd instructions of backup, which might result in the inconsistent states between primary and backup in following operations. So the interruptions are supposed to be delivered at the same point in the instruction streams.
- Multi-core
  - The solution in the paper of this lecture is not allowing multi-core to prevent the possible concurrency. For example, on primary there are 2 threads of different cores racing for the lock, and it should be guaranteed that the thread of backup that eventually wins the lock are the same as the thread of primary, which requires a whole bunch of machinery or complexities that are hard to deal with. Here assumes that the processor is unique.

## VM-FT interruptions handling

When received the interruption, VM-FT knows how many instructions CPU has executed (e.g. 100 instructions), and calculate a position (e.g. 100), inform the backup that when the instructions to execute the interruption after the 100th instruction. Most of the processors (e.g. x86) will stop at the required instruction and give back control to the operating system (here the VM monitor).

To arrange this, the backup needs to lag behind one message.

**The deterministic operations do not need to sync through logging channel as they have a copy of all the instructions.**

## VM-FT non-deterministic instructions handling

It firstly scans through all the non-deterministic instructions in Linux before starting the guest spaceor boot to the VM, and makes sure that they are transferred into invalid instructions. When guest space executing these non-deterministic instructions, it transfers control to the hypervisor through the trap, and the hypervisor will emulate to execute the instruction, and record the result to a certain register and send it back. The backup will execute this instruction and trap into kernel as well, and it will wait until primary syncs the result of the instruction, thus achieve the same result on this non-deterministic instruction.

## VM-FT failover handling

E.g. Primary had a counter as 10, and client sent a request `inc` to increase it to 11, and the primary did it but failed right before sending response to the client. If backup handled it now by sync to the instruction from logging channel, the `inc` operation did not executed by backup. If the client sent the `inc` request again, what it will get is 11 rather that 12.

To prevent the above situations, VM-FT made an **Output Rule**. Before primary sending response to client, it will send message to backup through logging channel. When backup received the message and did the same operations, it will send `ack` message to primary (similar to TCP). Only after primary received the `ack` and assured that backup can have the same state will it send response to client.

The output rule can be seen in any kind of replication systems (e.g. raft or zookeper).

**As it is a network request from client, though it is a deterministic instruction, it is needed to inform backup and wait for the `ack` message before respond to client.**

**It is possible for client to receive the response twice.** And it can be dealt with TCP.

## VM-FT Performance

**As most of VM-FT operations based on level of machine instructions or level of interruptions, the performance paid a hit.**

According to the statistics from the paper, the performance keeps a rate of 0.94~0.99 comparing with non-FT when running at primary/backup situations, while if the network transmit and receive with huge amount of data, the performance declines apparently with a decline rate of nearly 30\%. The reason might be that primary needs to send the data to the backup and wait until backup processed all the data can it sends response to the client.

This explains why poeple prefers to use RSM on application-level rather than instruction-level. However, they usually need to modify the application if use RSM on application-level like GFS.
