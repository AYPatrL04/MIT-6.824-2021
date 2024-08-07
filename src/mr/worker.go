package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
)

type KeyValue struct {
	Key   string
	Value string
}

type KIterator []KeyValue

func (a KIterator) Len() int {
	return len(a)
}
func (a KIterator) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a KIterator) Less(i, j int) bool {
	return a[i].Key < a[j].Key
}

func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	work := true
	flag := false
	for work {
		task := Task{}
		if flag {
			task.TaskType = Wait
			flag = false
		}
		if GetTask(&task) {
			switch task.TaskType {
			case Map:
				err := Mapper(mapf, &task)
				if err == nil {
					FinishTask(&task)
				}
			case Reduce:
				err := Reducer(reducef, &task)
				if err == nil {
					FinishTask(&task)
				}
			case Wait:
				time.Sleep(time.Second)
				flag = true
			default:
				if task.TaskType == Done {
					work = false
				}
			}
		} else {
			break
		}
	}
}

func GetTask(reply *Task) bool {
	args := Args{}
	if reply.TaskType == Wait {
		args.Status = Wait
	}
	return call("Master.AllocateTask", &args, reply)
}

func Mapper(mapf func(string, string) []KeyValue, task *Task) error {
	fileName := task.FileName
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("cannot open %v", fileName)
		return err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", fileName)
		return err
	}
	file.Close()
	kva := mapf(fileName, string(content))
	intermediate := make([][]KeyValue, task.NReduce)
	for _, kv := range kva {
		i := ihash(kv.Key) % task.NReduce
		intermediate[i] = append(intermediate[i], kv)
	}
	for i, kvs := range intermediate {
		oname := "mr-" + strconv.Itoa(task.MapId) + "-" + strconv.Itoa(i)
		f, err := os.Create(oname)
		if err != nil {
			log.Fatalf("cannot create %v", oname)
			return err
		}
		enc := json.NewEncoder(f)
		for _, kv := range kvs {
			enc.Encode(&kv)
		}
		f.Close()
	}
	return nil
}

func Reducer(reducef func(string, []string) string, task *Task) error {
	maps := make(map[string][]string)
	for i := 0; i < task.NMap; i++ {
		fileName := fmt.Sprintf("mr-%d-%d", i, task.ReduceId)
		file, _ := os.Open(fileName)
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			if _, ok := maps[kv.Key]; !ok {
				maps[kv.Key] = make([]string, 0, 100)
			}
			maps[kv.Key] = append(maps[kv.Key], kv.Value)
		}
	}
	oname := fmt.Sprintf("mr-out-%d", task.TaskId)
	res := make([]string, 0, 100)
	for k, v := range maps {
		res = append(res, fmt.Sprintf("%v %v\n", k, reducef(k, v)))
	}
	if err := ioutil.WriteFile(oname, []byte(strings.Join(res, "")), 0600); err != nil {
		log.Fatalf("fail to write %v", oname)
		return err
	}
	return nil
}

func FinishTask(task *Task) {
	reply := Reply{}
	call("Master.Finish", task, &reply)
}

func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		DPrintf("dialing: ", err)
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
