package kvraft

import (
	"MIT-6_824-2021/labgob"
	"MIT-6_824-2021/labrpc"
	"MIT-6_824-2021/raft"
	"bytes"
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

	lastIncludedIndex int
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	if kv.killed() {
		reply.Err = ErrWrongLeader
		return
	}
	op := Op{
		OpType:   "Get",
		Key:      args.Key,
		ClientId: args.ClientId,
		SeqId:    args.SeqId,
	}
	lastIdx, _, leader := kv.rf.Start(op)
	if !leader {
		reply.Err = ErrWrongLeader
		return
	}
	kv.mu.Lock()
	ch := kv.getChannel(lastIdx)
	kv.mu.Unlock()
	select {
	case replyMsg := <-ch:
		if replyMsg.ClientId == op.ClientId && replyMsg.SeqId == op.SeqId {
			reply.Value = replyMsg.Value
			reply.Err = OK
		} else {
			reply.Err = ErrWrongLeader
		}
	case <-time.After(100 * time.Millisecond):
		reply.Err = ErrWrongLeader
	}
	go func() {
		kv.mu.Lock()
		delete(kv.applyChMap, lastIdx)
		kv.mu.Unlock()
	}()
}

func (kv *KVServer) getChannel(index int) chan Op {
	ch, ok := kv.applyChMap[index]
	if !ok {
		kv.applyChMap[index] = make(chan Op, 1)
		ch = kv.applyChMap[index]
	}
	return ch
}

func (kv *KVServer) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	if kv.killed() {
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
	lastIdx, _, leader := kv.rf.Start(op)
	if !leader {
		reply.Err = ErrWrongLeader
		return
	}
	kv.mu.Lock()
	ch := kv.getChannel(lastIdx)
	kv.mu.Unlock()
	select {
	case replyMsg := <-ch:
		if replyMsg.ClientId == op.ClientId && replyMsg.SeqId == op.SeqId {
			reply.Err = OK
		} else {
			reply.Err = ErrWrongLeader
		}
	case <-time.After(100 * time.Millisecond):
		reply.Err = ErrWrongLeader
	}
	go func() {
		kv.mu.Lock()
		delete(kv.applyChMap, lastIdx)
		kv.mu.Unlock()
	}()
}

func (kv *KVServer) applyMsgHandler() {
	for !kv.killed() {
		select {
		case msg := <-kv.applyCh:
			if msg.CommandValid {
				kv.mu.Lock()
				idx := msg.CommandIndex
				if idx <= kv.lastIncludedIndex {
					kv.mu.Unlock()
					continue
				}
				op := msg.Command.(Op)
				kv.lastIncludedIndex = idx
				if op.OpType == "Get" {
					op.Value = kv.kvStore[op.Key]
				} else if !kv.duplicateOp(op.ClientId, op.SeqId) {
					switch op.OpType {
					case "Put":
						kv.kvStore[op.Key] = op.Value
					case "Append":
						kv.kvStore[op.Key] += op.Value
					}
					kv.clientSeqMap[op.ClientId] = op.SeqId
				}
				if _, leader := kv.rf.GetState(); leader {
					kv.getChannel(idx) <- op
				}
				if kv.maxraftstate != -1 && kv.rf.GetStateSize() > kv.maxraftstate {
					kv.rf.Snapshot(idx, kv.Persistence())
				}
				kv.mu.Unlock()
			}
			if msg.SnapshotValid {
				kv.mu.Lock()
				kv.DecodeSnapShot(msg.Snapshot)
				kv.lastIncludedIndex = msg.SnapshotIndex
				kv.mu.Unlock()
			}
		}
	}
}

func (kv *KVServer) Persistence() []byte {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(kv.kvStore)
	e.Encode(kv.clientSeqMap)
	return w.Bytes()
}

func (kv *KVServer) DecodeSnapShot(snapshot []byte) {
	if snapshot == nil || len(snapshot) == 0 {
		return
	}
	r := bytes.NewBuffer(snapshot)
	d := labgob.NewDecoder(r)
	var kvStore map[string]string
	var clientSeqMap map[int64]int
	if d.Decode(&kvStore) != nil || d.Decode(&clientSeqMap) != nil {
		log.Fatal("Decode snapshot error")
	} else {
		kv.kvStore = kvStore
		kv.clientSeqMap = clientSeqMap
	}
}

func (kv *KVServer) duplicateOp(clientId int64, seqId int) bool {
	if lastSeqId, ok := kv.clientSeqMap[clientId]; ok {
		return seqId <= lastSeqId
	}
	return false
}

func (kv *KVServer) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
}

func (kv *KVServer) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	labgob.Register(Op{})
	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate
	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)
	kv.clientSeqMap = make(map[int64]int)
	kv.kvStore = make(map[string]string)
	kv.applyChMap = make(map[int]chan Op)
	kv.lastIncludedIndex = -1
	snapshot := persister.ReadSnapshot()
	kv.mu.Lock()
	kv.DecodeSnapShot(snapshot)
	kv.mu.Unlock()
	go kv.applyMsgHandler()
	return kv
}
