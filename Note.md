<h1 align="center">MIT 6.824 Learning Notes</h1>
<h5 align="center">AYPatrL04</h5>
<h5 align="center">AYPatrL04@gmail.com</h5>

<h6 align="center">======= Lec.01 Fri. 02 Aug. 2024 =======</h6>

## Introduction

### Labs
1) MapReduce
2) Replication Using Raft
3) Replicated Key-Value Service
4) Sharded Key-Value Service

### Focus on infrastructure, not application.
- Storage
- Computation
- Communication

### Main topics:
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

The different machines running same task will produce the same output and write it in the intermediate files. The different processes will do atomic rename to make sure that only one process writes the final output.

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
- the process of marshal and unmarshal parameters happens in the stub.
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
