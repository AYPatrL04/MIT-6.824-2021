package mr

import (
	"fmt"
	"log"
	"os"
)
import "strconv"

type TaskPhase int

const (
	MapPhase    TaskPhase = 0
	ReducePhase TaskPhase = 1
)
const Debug = false

func DPrintf(format string, v ...interface{}) {
	if Debug {
		log.Printf(format+"\n", v...)
	}
}

type Task struct {
	FileName string
	NReduce  int
	NMaps    int
	Seq      int
	Phase    TaskPhase
	Alive    bool // worker should exit when alive is false
}

func reduceName(mapIdx, reduceIdx int) string {
	return fmt.Sprintf("mr-%d-%d", mapIdx, reduceIdx)
}

func mergeName(reduceIdx int) string {
	return fmt.Sprintf("mr-out-%d", reduceIdx)
}

type TaskArgs struct {
	WorkerId int
}

type TaskReply struct {
	Task *Task
}

type ReportTaskArgs struct {
	Done     bool
	Seq      int
	Phase    TaskPhase
	WorkerId int
}

type ReportTaskReply struct {
}

type RegisterArgs struct {
}

type RegisterReply struct {
	WorkerId int
}

func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
