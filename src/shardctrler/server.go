package shardctrler

import (
	"MIT-6_824-2021/labgob"
	"MIT-6_824-2021/labrpc"
	"MIT-6_824-2021/raft"
	"sort"
	"sync"
	"time"
)

type ShardCtrler struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg

	applyChMap   map[int]chan Op
	clientSeqMap map[int64]int

	configs []Config // indexed by config num
}

type Op struct {
	OpType   string
	ClientId int64
	SeqId    int

	JoinServers map[int][]string
	LeaveGIDs   []int
	MoveShard   int
	MoveGID     int
	QueryNum    int
}

func (sc *ShardCtrler) getChannel(index int) chan Op {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	ch, ok := sc.applyChMap[index]
	if !ok {
		sc.applyChMap[index] = make(chan Op, 1)
		ch = sc.applyChMap[index]
	}
	return ch
}

func (sc *ShardCtrler) Join(args *JoinArgs, reply *JoinReply) {
	reply.WrongLeader = false
	op := Op{
		OpType:      JOIN,
		ClientId:    args.ClientId,
		SeqId:       args.SeqNum,
		JoinServers: args.Servers,
	}
	lastIdx, _, leader := sc.rf.Start(op)
	if !leader {
		reply.WrongLeader = true
		return
	}
	ch := sc.getChannel(lastIdx)
	select {
	case replyMsg := <-ch:
		if replyMsg.ClientId == op.ClientId && replyMsg.SeqId == op.SeqId {
			reply.Err = OK
		} else {
			reply.WrongLeader = true
		}
	case <-time.After(100 * time.Millisecond):
		reply.WrongLeader = true
	}
	go func() {
		sc.mu.Lock()
		delete(sc.applyChMap, lastIdx)
		sc.mu.Unlock()
	}()
}

func (sc *ShardCtrler) Leave(args *LeaveArgs, reply *LeaveReply) {
	reply.WrongLeader = false
	op := Op{
		OpType:    LEAVE,
		ClientId:  args.ClientId,
		SeqId:     args.SeqNum,
		LeaveGIDs: args.GIDs,
	}
	lastIdx, _, leader := sc.rf.Start(op)
	if !leader {
		reply.WrongLeader = true
		return
	}
	ch := sc.getChannel(lastIdx)
	select {
	case replyMsg := <-ch:
		if replyMsg.ClientId == op.ClientId && replyMsg.SeqId == op.SeqId {
			reply.Err = OK
			return
		} else {
			reply.WrongLeader = true
		}
	case <-time.After(100 * time.Millisecond):
		reply.WrongLeader = true
	}
	go func() {
		sc.mu.Lock()
		delete(sc.applyChMap, lastIdx)
		sc.mu.Unlock()
	}()
}

func (sc *ShardCtrler) Move(args *MoveArgs, reply *MoveReply) {
	reply.WrongLeader = false
	op := Op{
		OpType:    MOVE,
		ClientId:  args.ClientId,
		SeqId:     args.SeqNum,
		MoveShard: args.Shard,
		MoveGID:   args.GID,
	}
	lastIdx, _, leader := sc.rf.Start(op)
	if !leader {
		reply.WrongLeader = true
		return
	}
	ch := sc.getChannel(lastIdx)
	select {
	case replyMsg := <-ch:
		if replyMsg.ClientId == op.ClientId && replyMsg.SeqId == op.SeqId {
			reply.Err = OK
			return
		} else {
			reply.WrongLeader = true
		}
	case <-time.After(100 * time.Millisecond):
		reply.WrongLeader = true
	}
	go func() {
		sc.mu.Lock()
		delete(sc.applyChMap, lastIdx)
		sc.mu.Unlock()
	}()
}

func (sc *ShardCtrler) Query(args *QueryArgs, reply *QueryReply) {
	reply.WrongLeader = false
	op := Op{
		OpType:   QUERY,
		ClientId: args.ClientId,
		SeqId:    args.SeqNum,
		QueryNum: args.Num,
	}
	lastIdx, _, leader := sc.rf.Start(op)
	if !leader {
		reply.WrongLeader = true
		return
	}
	ch := sc.getChannel(lastIdx)
	select {
	case replyMsg := <-ch:
		if replyMsg.ClientId == op.ClientId && replyMsg.SeqId == op.SeqId {
			sc.mu.Lock()
			reply.Err = OK
			sc.clientSeqMap[op.ClientId] = op.SeqId
			if op.QueryNum == -1 || op.QueryNum >= len(sc.configs) {
				reply.Config = sc.configs[len(sc.configs)-1]
			} else {
				reply.Config = sc.configs[op.QueryNum]
			}
			sc.mu.Unlock()
		} else {
			reply.WrongLeader = true
		}
	case <-time.After(100 * time.Millisecond):
		reply.WrongLeader = true
	}
	go func() {
		sc.mu.Lock()
		delete(sc.applyChMap, lastIdx)
		sc.mu.Unlock()
	}()
}

func (sc *ShardCtrler) applyMsgHandler() {
	for {
		select {
		case msg := <-sc.applyCh:
			if msg.CommandValid {
				idx := msg.CommandIndex
				op := msg.Command.(Op)
				sc.mu.Lock()
				if !sc.duplicateOp(op.ClientId, op.SeqId) {
					switch op.OpType {
					case JOIN:
						sc.clientSeqMap[op.ClientId] = op.SeqId
						sc.configs = append(sc.configs, *sc.joinHandler(op.JoinServers))
					case LEAVE:
						sc.clientSeqMap[op.ClientId] = op.SeqId
						sc.configs = append(sc.configs, *sc.leaveHandler(op.LeaveGIDs))
					case MOVE:
						sc.clientSeqMap[op.ClientId] = op.SeqId
						sc.configs = append(sc.configs, *sc.moveHandler(op.MoveShard, op.MoveGID))
					}
					sc.clientSeqMap[op.ClientId] = op.SeqId
				}
				sc.mu.Unlock()
				sc.getChannel(idx) <- op
			}
		}
	}
}

func (sc *ShardCtrler) joinHandler(servers map[int][]string) *Config {
	lastConfig := sc.configs[len(sc.configs)-1]
	newGroups := make(map[int][]string)
	for gid, serverLists := range lastConfig.Groups {
		newGroups[gid] = serverLists
	}
	for gid, serverLists := range servers {
		newGroups[gid] = serverLists
	}
	GroupMap := make(map[int]int)
	for gid := range newGroups {
		GroupMap[gid] = 0
	}
	for _, gid := range lastConfig.Shards {
		if gid != 0 {
			GroupMap[gid]++
		}
	}
	if len(GroupMap) == 0 {
		return &Config{
			Num:    len(sc.configs),
			Shards: [NShards]int{},
			Groups: newGroups,
		}
	}
	return &Config{
		Num:    len(sc.configs),
		Shards: sc.loadBalance(GroupMap, lastConfig.Shards),
		Groups: newGroups,
	}
}

func (sc *ShardCtrler) leaveHandler(gids []int) *Config {
	leaveGroups := make(map[int]bool)
	for _, gid := range gids {
		leaveGroups[gid] = true
	}
	lastConfig := sc.configs[len(sc.configs)-1]
	newGroups := make(map[int][]string)
	for gid, servers := range lastConfig.Groups {
		newGroups[gid] = servers
	}
	for _, gid := range gids {
		delete(newGroups, gid)
	}
	Groups := make(map[int]int)
	newShard := lastConfig.Shards
	for gid := range newGroups {
		if !leaveGroups[gid] {
			Groups[gid] = 0
		}
	}
	for shard, gid := range lastConfig.Shards {
		if gid != 0 {
			if leaveGroups[gid] {
				newShard[shard] = 0
			} else {
				Groups[gid]++
			}
		}
	}
	if len(Groups) == 0 {
		return &Config{
			Num:    len(sc.configs),
			Shards: [NShards]int{},
			Groups: newGroups,
		}
	}
	return &Config{
		Num:    len(sc.configs),
		Shards: sc.loadBalance(Groups, newShard),
		Groups: newGroups,
	}
}

func (sc *ShardCtrler) moveHandler(shard int, gid int) *Config {
	lastConfig := sc.configs[len(sc.configs)-1]
	newShards := lastConfig.Shards
	newShards[shard] = gid
	return &Config{
		Num:    len(sc.configs),
		Shards: newShards,
		Groups: lastConfig.Groups,
	}
}

func (sc *ShardCtrler) duplicateOp(clientId int64, seqId int) bool {
	if lastSeqId, ok := sc.clientSeqMap[clientId]; ok {
		return seqId <= lastSeqId
	}
	return false
}

func (sc *ShardCtrler) loadBalance(Groups map[int]int, shards [NShards]int) [NShards]int {
	length := len(Groups)
	average, remainder := NShards/length, NShards%length
	sortGids := sortGid(Groups)
	for i := 0; i < length; i++ {
		target := average
		if !moreAllocated(length, remainder, i) {
			target = average + 1
		}
		if Groups[sortGids[i]] > target {
			overloadGid := sortGids[i]
			diff := Groups[overloadGid] - target
			for shard, gid := range shards {
				if diff <= 0 {
					break
				}
				if gid == overloadGid {
					shards[shard] = InvalidGID
					diff--
				}
			}
			Groups[overloadGid] = target
		}
	}
	for i := 0; i < length; i++ {
		target := average
		if !moreAllocated(length, remainder, i) {
			target = average + 1
		}
		if Groups[sortGids[i]] < target {
			freeGid := sortGids[i]
			diff := target - Groups[freeGid]
			for shard, gid := range shards {
				if diff <= 0 {
					break
				}
				if gid == InvalidGID {
					shards[shard] = freeGid
					diff--
				}
			}
			Groups[freeGid] = target
		}
	}
	return shards
}

func sortGid(Groups map[int]int) []int {
	gids := make([]int, 0, len(Groups))
	for gid := range Groups {
		gids = append(gids, gid)
	}
	sort.Slice(gids, func(i, j int) bool {
		if Groups[gids[i]] == Groups[gids[j]] {
			return gids[i] < gids[j]
		} else {
			return Groups[gids[i]] < Groups[gids[j]]
		}
	})
	return gids
}

func moreAllocated(length, remainder, i int) bool {
	return i < length-remainder
}

func (sc *ShardCtrler) Kill() {
	sc.rf.Kill()
}

func (sc *ShardCtrler) Raft() *raft.Raft {
	return sc.rf
}

func StartServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister) *ShardCtrler {
	sc := new(ShardCtrler)
	sc.me = me
	sc.configs = make([]Config, 1)
	sc.configs[0].Groups = map[int][]string{}
	labgob.Register(Op{})
	sc.applyCh = make(chan raft.ApplyMsg)
	sc.rf = raft.Make(servers, me, persister, sc.applyCh)
	sc.applyChMap = make(map[int]chan Op)
	sc.clientSeqMap = make(map[int64]int)
	go sc.applyMsgHandler()
	return sc
}
