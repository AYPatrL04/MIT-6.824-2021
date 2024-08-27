package shardkv

import (
	"MIT-6_824-2021/labrpc"
	"MIT-6_824-2021/shardctrler"
	"bytes"
	"log"
	"sync/atomic"
	"time"
)
import "MIT-6_824-2021/raft"
import "sync"
import "MIT-6_824-2021/labgob"

type Op struct {
	ClientId     int64
	SeqNum       int
	OpType       string
	Key          string
	Value        string
	Shard        Shard
	ShardId      int
	UpdateConfig shardctrler.Config
	ClientSeqMap map[int64]int
}

type OpReply struct {
	Err      Err
	SeqNum   int
	ClientId int64
}

type ShardKV struct {
	mu           sync.Mutex
	me           int
	rf           *raft.Raft
	applyCh      chan raft.ApplyMsg
	make_end     func(string) *labrpc.ClientEnd
	gid          int
	ctrlers      []*labrpc.ClientEnd
	maxraftstate int // snapshot if log grows this big
	dead         int32
	Config       shardctrler.Config
	LastConfig   shardctrler.Config

	shards       []Shard
	applyChMap   map[int]chan OpReply
	clientSeqMap map[int64]int
	mck          *shardctrler.Clerk
}

type Shard struct {
	Data    map[string]string
	Version int
}

func (kv *ShardKV) getChannel(index int) chan OpReply {
	ch, ok := kv.applyChMap[index]
	if !ok {
		kv.applyChMap[index] = make(chan OpReply, 1)
		ch = kv.applyChMap[index]
	}
	return ch
}

func (kv *ShardKV) Get(args *GetArgs, reply *GetReply) {
	shardId := key2shard(args.Key)
	kv.mu.Lock()
	if kv.Config.Shards[shardId] != kv.gid {
		reply.Err = ErrWrongGroup
		kv.mu.Unlock()
		return
	} else if kv.shards[shardId].Data == nil {
		reply.Err = ShardNotReady
		kv.mu.Unlock()
		return
	}
	kv.mu.Unlock()
	op := Op{
		OpType:   GET,
		Key:      args.Key,
		ClientId: args.ClientId,
		SeqNum:   args.RequestId,
	}
	err := kv.DoOp(op, GetTimeout)
	if err != OK {
		reply.Err = err
		return
	}
	kv.mu.Lock()
	if kv.Config.Shards[shardId] != kv.gid {
		reply.Err = ErrWrongGroup
	} else if kv.shards[shardId].Data == nil {
		reply.Err = ShardNotReady
	} else {
		reply.Err = OK
		reply.Value = kv.shards[shardId].Data[args.Key]
	}
	kv.mu.Unlock()
}

func (kv *ShardKV) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	shardId := key2shard(args.Key)
	kv.mu.Lock()
	if kv.Config.Shards[shardId] != kv.gid {
		reply.Err = ErrWrongGroup
		kv.mu.Unlock()
		return
	} else if kv.shards[shardId].Data == nil {
		reply.Err = ShardNotReady
		kv.mu.Unlock()
		return
	}
	kv.mu.Unlock()
	op := Op{
		OpType:   args.Op,
		Key:      args.Key,
		Value:    args.Value,
		ClientId: args.ClientId,
		SeqNum:   args.RequestId,
	}
	reply.Err = kv.DoOp(op, PutAppendTimeout)
}

func (kv *ShardKV) AddShard(args *SendShardArgs, reply *SendShardReply) {
	op := Op{
		OpType:       ADDSHARD,
		Shard:        args.Shard,
		ShardId:      args.ShardId,
		ClientId:     args.ClientId,
		SeqNum:       args.RequestId,
		ClientSeqMap: args.LastApplied,
	}
	reply.Err = kv.DoOp(op, AddShardsTimeout)
}

func (kv *ShardKV) DoOp(op Op, timeout time.Duration) Err {
	kv.mu.Lock()
	lastIdx, _, leader := kv.rf.Start(op)
	if !leader {
		kv.mu.Unlock()
		return ErrWrongLeader
	}
	ch := kv.getChannel(lastIdx)
	kv.mu.Unlock()
	select {
	case reply := <-ch:
		kv.mu.Lock()
		delete(kv.applyChMap, lastIdx)
		if reply.ClientId != op.ClientId || reply.SeqNum != op.SeqNum {
			kv.mu.Unlock()
			return ErrInconsistentData
		}
		kv.mu.Unlock()
		return reply.Err
	case <-time.After(timeout):
		return ErrTimeout
	}
}

func (kv *ShardKV) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
}

func (kv *ShardKV) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

func (kv *ShardKV) applyMsgHandler() {
	for {
		if kv.killed() {
			return
		}
		select {
		case msg := <-kv.applyCh:
			if msg.CommandValid {
				kv.mu.Lock()
				op := msg.Command.(Op)
				reply := OpReply{
					SeqNum:   op.SeqNum,
					ClientId: op.ClientId,
					Err:      OK,
				}
				if isKVOp(op.OpType) {
					shardId := key2shard(op.Key)
					if kv.Config.Shards[shardId] != kv.gid {
						reply.Err = ErrWrongGroup
					} else if kv.shards[shardId].Data == nil {
						reply.Err = ShardNotReady
					} else {
						if !kv.duplicate(op.ClientId, op.SeqNum) {
							kv.clientSeqMap[op.ClientId] = op.SeqNum
							switch op.OpType {
							case PUT:
								kv.shards[shardId].Data[op.Key] = op.Value
							case APPEND:
								kv.shards[shardId].Data[op.Key] += op.Value
							case GET:
							default:
								kv.mu.Unlock()
								log.Fatalf("Invalid Operation Type: %v.\n", op.OpType)
							}
						}
					}
				} else {
					switch op.OpType {
					case CONFIG:
						kv.updateConfig(op)
					case ADDSHARD:
						if kv.Config.Num < op.SeqNum {
							reply.Err = ConfigNotReady
							break
						}
						kv.addShard(op)
					case DELSHARD:
						kv.removeShard(op)
					default:
						kv.mu.Unlock()
						log.Fatalf("Invalid Operation Type: %v.\n", op.OpType)
					}
				}
				if kv.maxraftstate != -1 && kv.rf.GetStateSize() > kv.maxraftstate {
					kv.rf.Snapshot(msg.CommandIndex, kv.Persist())
				}
				ch := kv.getChannel(msg.CommandIndex)
				kv.mu.Unlock()
				ch <- reply
			}
			if msg.SnapshotValid {
				if kv.rf.CondInstallSnapshot(msg.SnapshotTerm, msg.SnapshotIndex, msg.Snapshot) {
					kv.mu.Lock()
					kv.DecodeSnapshot(msg.Snapshot)
					kv.mu.Unlock()
				}
				continue
			}
		}
	}
}

func isKVOp(op string) bool {
	return op == PUT || op == APPEND || op == GET
}

func (kv *ShardKV) duplicate(clientId int64, seqNum int) bool {
	lastSeq, ok := kv.clientSeqMap[clientId]
	return ok && lastSeq >= seqNum
}

func (kv *ShardKV) updateConfig(op Op) {
	curConfig := kv.Config
	newConfig := op.UpdateConfig
	if curConfig.Num >= newConfig.Num {
		return
	}
	for shardId, gid := range newConfig.Shards {
		if gid == kv.gid && curConfig.Shards[shardId] == 0 {
			kv.shards[shardId].Data = make(map[string]string)
			kv.shards[shardId].Version = newConfig.Num
		}
	}
	kv.LastConfig = curConfig
	kv.Config = newConfig
}

func (kv *ShardKV) addShard(op Op) {
	if op.Shard.Version < kv.Config.Num ||
		kv.shards[op.ShardId].Data != nil {
		return
	}
	kv.shards[op.ShardId] = kv.clone(op.Shard.Version, op.Shard.Data)
	for client, seq := range op.ClientSeqMap {
		if record, ok := kv.clientSeqMap[client]; !ok || record < seq {
			kv.clientSeqMap[client] = seq
		}
	}
}

func (kv *ShardKV) removeShard(op Op) {
	if op.SeqNum < kv.Config.Num {
		return
	}
	kv.shards[op.ShardId].Data = nil
	kv.shards[op.ShardId].Version = op.SeqNum
}

func (kv *ShardKV) clone(ConfigNum int, KVMap map[string]string) Shard {
	newShard := Shard{
		Data:    make(map[string]string),
		Version: ConfigNum,
	}
	for k, v := range KVMap {
		newShard.Data[k] = v
	}
	return newShard
}

func (kv *ShardKV) Persist() []byte {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(kv.Config)
	e.Encode(kv.LastConfig)
	e.Encode(kv.shards)
	e.Encode(kv.clientSeqMap)
	e.Encode(kv.maxraftstate)
	return w.Bytes()
}

func (kv *ShardKV) DecodeSnapshot(snapshot []byte) {
	if snapshot == nil || len(snapshot) < 1 {
		return
	}
	r := bytes.NewBuffer(snapshot)
	d := labgob.NewDecoder(r)
	var Config shardctrler.Config
	var LastConfig shardctrler.Config
	var shards []Shard
	var clientSeqMap map[int64]int
	var maxraftstate int
	if d.Decode(&Config) != nil || d.Decode(&LastConfig) != nil || d.Decode(&shards) != nil || d.Decode(&clientSeqMap) != nil || d.Decode(&maxraftstate) != nil {
		log.Fatalf("DecodeSnapshot failed.\n")
	} else {
		kv.Config = Config
		kv.LastConfig = LastConfig
		kv.shards = shards
		kv.clientSeqMap = clientSeqMap
		kv.maxraftstate = maxraftstate
	}
}

func (kv *ShardKV) ConfigDetected() {
	kv.mu.Lock()
	curConfig := kv.Config
	rf := kv.rf
	kv.mu.Unlock()
	for !kv.killed() {
		if _, leader := rf.GetState(); !leader {
			time.Sleep(UpdateConfigInterval)
			continue
		}
		kv.mu.Lock()
		if !kv.allSent() {
			clientSeqMap := make(map[int64]int)
			for client, seq := range kv.clientSeqMap {
				clientSeqMap[client] = seq
			}
			for shardId, gid := range kv.LastConfig.Shards {
				if gid == kv.gid &&
					kv.Config.Shards[shardId] != kv.gid &&
					kv.shards[shardId].Version < kv.Config.Num {
					sendShard := kv.clone(kv.Config.Num, kv.shards[shardId].Data)
					args := SendShardArgs{
						Shard:       sendShard,
						ShardId:     shardId,
						ClientId:    int64(gid),
						RequestId:   kv.Config.Num,
						LastApplied: clientSeqMap,
					}
					serverList := kv.Config.Groups[kv.Config.Shards[shardId]]
					servers := make([]*labrpc.ClientEnd, len(serverList))
					for i, server := range serverList {
						servers[i] = kv.make_end(server)
					}
					go func(servers []*labrpc.ClientEnd, args *SendShardArgs) {
						serverIdx := 0
						start := time.Now()
						for {
							var reply SendShardReply
							ok := servers[serverIdx].Call("ShardKV.AddShard", args, &reply)
							if ok && reply.Err == OK || time.Now().Sub(start) >= 2000*time.Millisecond {
								kv.mu.Lock()
								op := Op{
									OpType:   DELSHARD,
									ShardId:  args.ShardId,
									ClientId: int64(kv.gid),
									SeqNum:   kv.Config.Num,
								}
								kv.mu.Unlock()
								kv.DoOp(op, DelShardsTimeout)
								break
							}
							serverIdx = (serverIdx + 1) % len(servers)
							if serverIdx == 0 {
								time.Sleep(UpdateConfigInterval)
							}
						}
					}(servers, &args)
				}
			}
			kv.mu.Unlock()
			time.Sleep(UpdateConfigInterval)
			continue
		}
		if !kv.allReceived() {
			kv.mu.Unlock()
			time.Sleep(UpdateConfigInterval)
			continue
		}
		curConfig = kv.Config
		mck := kv.mck
		kv.mu.Unlock()
		newConfig := mck.Query(curConfig.Num + 1)
		if newConfig.Num != curConfig.Num+1 {
			time.Sleep(UpdateConfigInterval)
			continue
		}
		op := Op{
			OpType:       CONFIG,
			UpdateConfig: newConfig,
			ClientId:     int64(kv.gid),
			SeqNum:       newConfig.Num,
		}
		kv.DoOp(op, ConfigTimeout)
	}
}

func (kv *ShardKV) allSent() bool {
	for shardId, shard := range kv.LastConfig.Shards {
		if shard == kv.gid &&
			kv.Config.Shards[shardId] != kv.gid &&
			kv.shards[shardId].Version < kv.Config.Num {
			return false
		}
	}
	return true
}

func (kv *ShardKV) allReceived() bool {
	for shardId, shard := range kv.LastConfig.Shards {
		if shard != kv.gid &&
			kv.Config.Shards[shardId] == kv.gid &&
			kv.shards[shardId].Version < kv.Config.Num {
			return false
		}
	}
	return true
}

func StartServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int, gid int, ctrlers []*labrpc.ClientEnd, make_end func(string) *labrpc.ClientEnd) *ShardKV {
	labgob.Register(Op{})
	kv := new(ShardKV)
	kv.me = me
	kv.maxraftstate = maxraftstate
	kv.make_end = make_end
	kv.gid = gid
	kv.ctrlers = ctrlers
	kv.clientSeqMap = make(map[int64]int)
	kv.mck = shardctrler.MakeClerk(kv.ctrlers)
	kv.applyChMap = make(map[int]chan OpReply)
	kv.shards = make([]Shard, shardctrler.NShards)
	snapshot := persister.ReadSnapshot()
	if len(snapshot) > 0 {
		kv.DecodeSnapshot(snapshot)
	}
	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)
	go kv.applyMsgHandler()
	go kv.ConfigDetected()
	return kv
}
