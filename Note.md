<h1 align="center">MIT 6.824 Learning Notes</h1>
<h5 align="center">AYPatrL04</h5>
<h5 align="center">AYPatrL04@gmail.com</h5>

<h6 align="center">======= Lec.1 Fri. 02 Aug. 2024 =======</h6>

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

## Abstract view
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