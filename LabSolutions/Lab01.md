# Solution for Lab 01

## Struct

```go
const (                          // Status code
	Done   = 0
	Map    = 1
	Reduce = 2
	Wait   = 3
)

type Master struct {             // Master(Coordinator)

	Status           int         // current work status of whole system

	MapTasks         []*Task     // info & status for each map task
	MapChan          chan *Task  // channel for map tasks
	MapTaskStatus    bool        // true: map finished

	ReduceTasks      []*Task     // info & status for each reduce task
	ReduceChan       chan *Task  // channel for reduce tasks
	ReduceTaskStatus bool        // true: reduce finished

	AvailableTaskCnt int         // refers to the unfinished task count 
                                 // renew when init tasks
}

type Args struct {               // definition for only status arguments
	Status int
}

type Reply struct {              // definition for only status replies
	Status int
}

type Task struct {               // definition for a single task
	TaskId    int                // identifier of the task
	TaskType  int                // task status
	FileName  string             // used to open file for map tasks
	NMap      int                // the total count for map tasks 
	NReduce   int                // the total count for reduce tasks
	Finished  bool               // remark if the task has finished
	MapId     int                // using for map tasks to concat the file name
	ReduceId  int                // using for reduce tasks to concat the file name
	timeStamp time.Time          // using to prevent worker crash
}
```

## Process

1. The `Master` was firstly created with every task marked `Map` and put all tasks into the `Map Channel`.
2. The `Worker` started asking for work, using a flag `work := true` and a loop `for work {}` to prevent the early exit.
3. `Worker` make a call to the `AllocateTask` function in `Master`.
4. `Master` check if all map tasks has been finished, and if not, it delivers a map task to the worker.
5. The `Worker` receives the work, by identifying the `TaskType` to decide which functions needed to be called. Here it is `Mapper`, and `FinishTask` if no error occurred.
6. The `Master` receives the detail of the finished task, remark the task to `Finished = true`, and update the `AvailableTaskCnt`. After that, it checks if all map tasks have been finished, and if yes, the `Status` of `Master` turned to `Reduce`, with `Reduce Channel` filled with tasks and all related status updated.
7. `Workers` do reduce tasks, the process is similar to map task.
8. After all works allocated and done, `AllocateTask` function replies a `Done` status to `worker`, and the flag `work` turned to false, remarks the end of the whole work. Meanwhile, the `Done` function returns `true` if both `MapTaskStatus` and `ReduceTaskStatus` are marked as true.
9. The dial error usually happens as the `Master.Done` returns `true` before all workers stop asking for work.
