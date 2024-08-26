package shardctrler

//
// Shardctrler clerk.
//

import (
	"MIT-6_824-2021/labrpc"
	rand2 "math/rand"
	"time"
)
import "crypto/rand"
import "math/big"

type Clerk struct {
	servers  []*labrpc.ClientEnd
	ClientId int64
	leaderId int
	SeqNum   int
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	ck.ClientId = nrand()
	ck.leaderId = rand2.Intn(len(ck.servers))
	return ck
}

func (ck *Clerk) Query(num int) Config {
	ck.SeqNum++
	leaderId := ck.leaderId
	args := &QueryArgs{
		Num:      num,
		ClientId: ck.ClientId,
		SeqNum:   ck.SeqNum,
	}
	for {
		reply := &QueryReply{}
		ok := ck.servers[leaderId].Call("ShardCtrler.Query", args, reply)
		if ok {
			if reply.Err == OK {
				ck.leaderId = leaderId
				return reply.Config
			} else if reply.WrongLeader {
				leaderId = (leaderId + 1) % len(ck.servers)
				continue
			}
		}
		leaderId = (leaderId + 1) % len(ck.servers)
		time.Sleep(100 * time.Millisecond)
	}
}

func (ck *Clerk) Join(servers map[int][]string) {
	ck.SeqNum++
	leaderId := ck.leaderId
	args := &JoinArgs{
		Servers:  servers,
		ClientId: ck.ClientId,
		SeqNum:   ck.SeqNum,
	}
	for {
		reply := &JoinReply{}
		ok := ck.servers[leaderId].Call("ShardCtrler.Join", args, reply)
		if ok {
			if reply.Err == OK {
				ck.leaderId = leaderId
				return
			} else if reply.WrongLeader {
				leaderId = (leaderId + 1) % len(ck.servers)
				continue
			}
		}
		leaderId = (leaderId + 1) % len(ck.servers)
		time.Sleep(100 * time.Millisecond)
	}
}

func (ck *Clerk) Leave(gids []int) {
	ck.SeqNum++
	leaderId := ck.leaderId
	args := &LeaveArgs{
		GIDs:     gids,
		SeqNum:   ck.SeqNum,
		ClientId: ck.ClientId,
	}
	for {
		reply := &LeaveReply{}
		ok := ck.servers[leaderId].Call("ShardCtrler.Leave", args, reply)
		if ok {
			if reply.Err == OK {
				ck.leaderId = leaderId
				return
			} else if reply.WrongLeader {
				leaderId = (leaderId + 1) % len(ck.servers)
				continue
			}
		}
		leaderId = (leaderId + 1) % len(ck.servers)
		time.Sleep(100 * time.Millisecond)
	}
}

func (ck *Clerk) Move(shard int, gid int) {
	ck.SeqNum++
	leaderId := ck.leaderId
	args := &MoveArgs{
		Shard:    shard,
		GID:      gid,
		ClientId: ck.ClientId,
		SeqNum:   ck.SeqNum,
	}
	for {
		reply := &MoveReply{}
		ok := ck.servers[leaderId].Call("ShardCtrler.Move", args, reply)
		if ok {
			if reply.Err == OK {
				ck.leaderId = leaderId
				return
			} else if reply.WrongLeader {
				leaderId = (leaderId + 1) % len(ck.servers)
				continue
			}
		}
		leaderId = (leaderId + 1) % len(ck.servers)
		time.Sleep(100 * time.Millisecond)
	}
}
