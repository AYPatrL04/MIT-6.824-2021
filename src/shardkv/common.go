package shardkv

import "time"

//
// Sharded key/value server.
// Lots of replica groups, each running Raft.
// Shardctrler decides which group serves each shard.
// Shardctrler may change shard assignment from time to time.
//
// You will have to modify these definitions.
//

const (
	OK                  = "OK"
	ErrNoKey            = "ErrNoKey"
	ErrWrongGroup       = "ErrWrongGroup"
	ErrWrongLeader      = "ErrWrongLeader"
	ShardNotReady       = "ShardNotReady"
	ConfigNotReady      = "ConfigNotReady"
	ErrInconsistentData = "ErrInconsistentData"
	ErrTimeout          = "ErrTimeout"

	PUT      = "Put"
	APPEND   = "Append"
	GET      = "Get"
	CONFIG   = "Config"
	ADDSHARD = "AddShard"
	DELSHARD = "DelShard"

	GetTimeout           = 500 * time.Millisecond
	PutAppendTimeout     = 500 * time.Millisecond
	AddShardsTimeout     = 500 * time.Millisecond
	DelShardsTimeout     = 500 * time.Millisecond
	ConfigTimeout        = 500 * time.Millisecond
	UpdateConfigInterval = 100 * time.Millisecond
)

type Err string

type PutAppendArgs struct {
	Key       string
	Value     string
	Op        string // "Put" or "Append"
	ClientId  int64
	RequestId int
}

type PutAppendReply struct {
	Err Err
}

type GetArgs struct {
	Key       string
	ClientId  int64
	RequestId int
}

type GetReply struct {
	Err   Err
	Value string
}

type SendShardArgs struct {
	ShardId     int
	Shard       Shard
	ClientId    int64
	RequestId   int
	LastApplied map[int64]int
}

type SendShardReply struct {
	Err Err
}
