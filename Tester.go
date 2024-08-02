package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strings"
	"sync"
)

func main() {
	//go reducer("a", []Pair{{"a", 1}, {"a", 1}, {"a", 1}})

	Serial("https://golang.org/", fetcher, make(map[string]bool))
	ConcurrentMutex("https://golang.org/", fetcher, makeState())
	ConcurrentChannel("https://golang.org/", fetcher)

}

// Lec01

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

// Lec02

type fetchState struct {
	mu      sync.Mutex
	fetched map[string]bool
}

type Fetcher interface {
	// Fetch returns a slice of URLs found on the page.
	Fetch(url string) (urls []string, err error)
}

type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) ([]string, error) {
	if res, ok := f[url]; ok {
		fmt.Printf("found: %s\n", url)
		return res.urls, nil
	}
	return nil, fmt.Errorf("not found: %s", url)
}

var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
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

func worker(url string, ch chan []string, fetcher Fetcher) {
	urls, err := fetcher.Fetch(url)
	if err != nil {
		ch <- []string{}
	} else {
		ch <- urls
	}
}

func coordinator(ch chan []string, fetcher Fetcher) {
	n := 1
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

func makeState() *fetchState {
	return &fetchState{fetched: make(map[string]bool)}
}

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
