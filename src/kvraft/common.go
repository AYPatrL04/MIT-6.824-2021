package kvraft

const (
	OK             = "OK"
	ErrNoKey       = "ErrNoKey"
	ErrWrongLeader = "ErrWrongLeader"
	//Get            = "Get"
	//Put            = "Put"
	//Append         = "Append"
)

type Err string

type PutAppendArgs struct {
	Key      string
	Value    string
	Op       string // "Put" or "Append"
	ClientId int64
	SeqId    int
}

type PutAppendReply struct {
	Err Err
}

type GetArgs struct {
	Key      string
	ClientId int64
	SeqId    int
}

type GetReply struct {
	Err   Err
	Value string
}
