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
package MIT_6_824

import "strings"

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

### Crawler

#### Serial

```go
package MIT_6_824

import "fmt"

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
package MIT_6_824

import (
    "fmt"
    "sync"
)

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
package MIT_6_824

import (
    "fmt"
)

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

[//]: # (generate a table with only 2 columns and 1 row using html)
<table>
  <tr>
    <td>Client</td>
    <td>Server</td>
  </tr>
  <tr>
    <td>
      <pre>
        <code>
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
func connect() *rpc.Client {
    client, err := rpc.DialHTTP("tcp", "localhost:1234")
    if err != nil {
        log.Fatal("dialing:", err)
    }
    return client
}
        </code>
      </pre>
    </td>
    <td>
      <pre>
        <code>
const (
	OK       = "OK"
	ErrNoKey = "ErrNoKey"
)
type Err string
type KV struct {
    mu   sync.Mutex
    data map[string]string
}
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
