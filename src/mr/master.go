package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type Master struct {
	Status           int
	MapTasks         []*Task
	MapChan          chan *Task
	MapTaskStatus    bool
	ReduceTasks      []*Task
	ReduceChan       chan *Task
	ReduceTaskStatus bool
	AvailableTaskCnt int
}

func (m *Master) AllocateTask(args *Args, reply *Task) error {
	mu.Lock()
	defer mu.Unlock()
	if m.Status == Map {
		if len(m.MapChan) > 0 {
			*reply = *<-m.MapChan
			m.MapTasks[reply.TaskId].timeStamp = time.Now()
		} else {
			if args.Status != Wait {
				reply.TaskType = Wait
			} else {
				for i := 0; i < len(m.MapTasks); i++ {
					if (!m.MapTasks[i].Finished) &&
						time.Since(m.MapTasks[i].timeStamp) > time.Second*10 {
						*reply = *m.MapTasks[i]
						m.MapTasks[i].timeStamp = time.Now()
						break
					}
				}
			}
		}
	} else if m.Status == Reduce {
		if len(m.ReduceChan) > 0 {
			*reply = *<-m.ReduceChan
			m.ReduceTasks[reply.TaskId].timeStamp = time.Now()
		} else {
			if args.Status != Wait {
				reply.TaskType = Wait
			} else {
				for i := 0; i < len(m.ReduceTasks); i++ {
					if (!m.ReduceTasks[i].Finished) &&
						time.Since(m.ReduceTasks[i].timeStamp) > time.Second*10 {
						*reply = *m.ReduceTasks[i]
						m.MapTasks[i].timeStamp = time.Now()
						break
					}
				}
			}
		}
	} else if m.MapTaskStatus && m.ReduceTaskStatus {
		reply.TaskType = Done
	} else {
		reply.TaskType = Wait
	}
	return nil
}

func (m *Master) Finish(task *Task, reply *Reply) error {
	if task.TaskType == Map {
		mu.Lock()
		if !m.MapTasks[task.TaskId].Finished {
			m.MapTasks[task.TaskId].Finished = true
			m.AvailableTaskCnt -= 1
			if m.AvailableTaskCnt == 0 {
				m.MapTaskStatus = true
			}
		}
		mu.Unlock()
	} else if task.TaskType == Reduce {
		mu.Lock()
		if !m.ReduceTasks[task.TaskId].Finished {
			m.ReduceTasks[task.TaskId].Finished = true
			m.AvailableTaskCnt -= 1
			if m.AvailableTaskCnt == 0 {
				m.ReduceTaskStatus = true
			}
		}
		mu.Unlock()
	}
	if m.Check(task.NReduce) {
		m.Status = Wait
	}
	return nil
}

func (m *Master) Check(NReduce int) bool {
	mu.Lock()
	defer mu.Unlock()
	if m.AvailableTaskCnt == 0 {
		if !m.ReduceTaskStatus && len(m.MapChan) == 0 {
			m.CreateReduceTask(NReduce)
			return false
		}
		return true
	}
	return false
}

func (m *Master) CreateReduceTask(NReduce int) {
	for i := 0; i < NReduce; i++ {
		task := Task{
			TaskType:  Reduce,
			FileName:  "mr-",
			NMap:      len(m.MapTasks),
			TaskId:    i,
			Finished:  false,
			ReduceId:  i,
			timeStamp: time.Now(),
		}
		m.ReduceChan <- &task
		m.ReduceTasks = append(m.ReduceTasks, &task)
	}
	m.Status = Reduce
	m.AvailableTaskCnt = NReduce
}

func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	// l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (m *Master) Done() bool {
	mu.Lock()
	defer mu.Unlock()
	return m.MapTaskStatus && m.ReduceTaskStatus
}

func MakeMaster(files []string, NReduce int) *Master {
	mu.Lock()
	defer mu.Unlock()
	m := Master{
		Status:           Map,
		MapChan:          make(chan *Task, len(files)),
		ReduceChan:       make(chan *Task, NReduce),
		MapTasks:         make([]*Task, len(files)),
		ReduceTasks:      make([]*Task, 0),
		MapTaskStatus:    false,
		ReduceTaskStatus: false,
		AvailableTaskCnt: len(files),
	}
	i := 0
	for _, file := range files {
		task := Task{
			TaskType:  Map,
			FileName:  file,
			NReduce:   NReduce,
			TaskId:    i,
			Finished:  false,
			MapId:     i,
			timeStamp: time.Now(),
		}
		m.MapChan <- &task
		m.MapTasks[i] = &task
		i++
	}
	m.server()
	return &m
}
