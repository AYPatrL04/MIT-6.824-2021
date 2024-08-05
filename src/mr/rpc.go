package mr

import (
	"os"
	"strconv"
	"sync"
	"time"
)

var mu sync.Mutex

type Args struct {
	Status int
}

type Reply struct {
	Status int
}

const (
	Done   = 0
	Map    = 1
	Reduce = 2
	Wait   = 3
)

type Task struct {
	TaskId    int
	TaskType  int
	FileName  string
	NMap      int
	NReduce   int
	Finished  bool
	MapId     int
	ReduceId  int
	timeStamp time.Time
}

func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
