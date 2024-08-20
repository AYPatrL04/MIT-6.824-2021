package kvraft

import (
	"MIT-6_824-2021/labgob"
	"MIT-6_824-2021/labrpc"
	"MIT-6_824-2021/raft"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type Op struct {
	OpType   string
	Key      string
	Value    string
	ClientId int64
	SeqId    int
	LogIdx   int
}

type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg
	dead    int32 // set by Kill()

	maxraftstate int // snapshot if log grows this big

	clientSeqMap map[int64]int
	kvStore      map[string]string
	applyChMap   map[int]chan Op

	//lastIncludedIndex int
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	if kv.killed() {
		reply.Err = ErrWrongLeader
		return
	}
	_, Leader := kv.rf.GetState()
	if !Leader {
		reply.Err = ErrWrongLeader
		return
	}
	op := Op{
		OpType:   "Get",
		Key:      args.Key,
		ClientId: args.ClientId,
		SeqId:    args.SeqId,
	}
	lastIdx, _, _ := kv.rf.Start(op)
	ch := kv.getChannel(lastIdx)
	defer func() {
		kv.mu.Lock()
		delete(kv.applyChMap, op.LogIdx)
		kv.mu.Unlock()
	}()
	timeout := time.NewTicker(100 * time.Millisecond)
	defer timeout.Stop()
	select {
	case replyMsg := <-ch:
		if replyMsg.ClientId == op.ClientId && replyMsg.SeqId == op.SeqId {
			kv.mu.Lock()
			reply.Value = kv.kvStore[args.Key]
			kv.mu.Unlock()
			reply.Err = OK
			return
		} else {
			reply.Err = ErrWrongLeader
		}
	case <-timeout.C:
		reply.Err = ErrWrongLeader
	}
}

func (kv *KVServer) getChannel(index int) chan Op {
	kv.mu.Lock()
	if _, ok := kv.applyChMap[index]; !ok {
		kv.applyChMap[index] = make(chan Op, 1)
	}
	ch := kv.applyChMap[index]
	kv.mu.Unlock()
	return ch
}

func (kv *KVServer) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	if kv.killed() {
		reply.Err = ErrWrongLeader
		return
	}
	_, Leader := kv.rf.GetState()
	if !Leader {
		reply.Err = ErrWrongLeader
		return
	}
	op := Op{
		OpType:   args.Op,
		Key:      args.Key,
		Value:    args.Value,
		ClientId: args.ClientId,
		SeqId:    args.SeqId,
	}
	lastIdx, _, _ := kv.rf.Start(op)
	ch := kv.getChannel(lastIdx)
	defer func() {
		kv.mu.Lock()
		delete(kv.applyChMap, lastIdx)
		kv.mu.Unlock()
	}()
	timeout := time.NewTicker(100 * time.Millisecond)
	select {
	case replyMsg := <-ch:
		if replyMsg.ClientId == op.ClientId && replyMsg.SeqId == op.SeqId {
			reply.Err = OK
			return
		} else {
			reply.Err = ErrWrongLeader
		}
	case <-timeout.C:
		reply.Err = ErrWrongLeader
	}
	defer timeout.Stop()
}

func (kv *KVServer) applyMsgHandler() {
	for {
		if kv.killed() {
			return
		}
		select {
		case msg := <-kv.applyCh:
			idx := msg.CommandIndex
			op := msg.Command.(Op)
			if !kv.duplicateOp(op.ClientId, op.SeqId) {
				kv.mu.Lock()
				switch op.OpType {
				case "Put":
					kv.kvStore[op.Key] = op.Value
				case "Append":
					kv.kvStore[op.Key] += op.Value
				}
				kv.clientSeqMap[op.ClientId] = op.SeqId
				kv.mu.Unlock()
			}
			kv.getChannel(idx) <- op
		}
	}
}

func (kv *KVServer) duplicateOp(clientId int64, seqId int) bool {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if lastSeqId, ok := kv.clientSeqMap[clientId]; ok {
		return seqId <= lastSeqId
	}
	return false
}

// the tester calls Kill() when a KVServer instance won't
// be needed again. for your convenience, we supply
// code to set rf.dead (without needing a lock),
// and a killed() method to test rf.dead in
// long-running loops. you can also add your own
// code to Kill(). you're not required to do anything
// about this, but it may be convenient (for example)
// to suppress debug output from a Kill()ed instance.
func (kv *KVServer) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
	// Your code here, if desired.
}

func (kv *KVServer) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// the k/v server should store snapshots through the underlying Raft
// implementation, which should call persister.SaveStateAndSnapshot() to
// atomically save the Raft state along with the snapshot.
// the k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
// in order to allow Raft to garbage-collect its log. if maxraftstate is -1,
// you don't need to snapshot.
// StartKVServer() must return quickly, so it should start goroutines
// for any long-running work.
func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	labgob.Register(Op{})

	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate

	// You may need initialization code here.

	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)

	// You may need initialization code here.
	kv.clientSeqMap = make(map[int64]int)
	kv.kvStore = make(map[string]string)
	kv.applyChMap = make(map[int]chan Op)
	//kv.lastIncludedIndex = -1
	//snapshot := persister.ReadSnapshot()
	//if len(snapshot) > 0 {
	//	kv.DecodeSnapShot(snapshot)
	//}
	go kv.applyMsgHandler()
	return kv
}
