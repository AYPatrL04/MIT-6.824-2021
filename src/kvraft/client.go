package kvraft

import (
	"MIT-6_824-2021/labrpc"
	"crypto/rand"
	"math/big"
	rand2 "math/rand"
)

type Clerk struct {
	servers  []*labrpc.ClientEnd
	LeaderId int
	ClientId int64
	SeqId    int
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
	ck.LeaderId = rand2.Intn(len(servers))
	ck.ClientId = nrand()
	ck.SeqId = 0
	return ck
}

func (ck *Clerk) Get(key string) string {
	ck.SeqId++
	args := GetArgs{Key: key, ClientId: ck.ClientId, SeqId: ck.SeqId}
	serverId := ck.LeaderId
	for {
		reply := GetReply{}
		ok := ck.servers[serverId].Call("KVServer.Get", &args, &reply)
		if ok {
			switch reply.Err {
			case OK:
				ck.LeaderId = serverId
				return reply.Value
			case ErrNoKey:
				ck.LeaderId = serverId
				return ""
			case ErrWrongLeader:
				serverId = (serverId + 1) % len(ck.servers)
				continue
			}
		}
		serverId = (serverId + 1) % len(ck.servers)
	}
}

func (ck *Clerk) PutAppend(key string, value string, op string) {
	ck.SeqId++
	args := PutAppendArgs{Key: key, Value: value, Op: op, ClientId: ck.ClientId, SeqId: ck.SeqId}
	serverId := ck.LeaderId
	for {
		reply := PutAppendReply{}
		ok := ck.servers[serverId].Call("KVServer.PutAppend", &args, &reply)
		if ok {
			switch reply.Err {
			case OK:
				ck.LeaderId = serverId
				return
			case ErrWrongLeader:
				serverId = (serverId + 1) % len(ck.servers)
				continue
			}
		}
		serverId = (serverId + 1) % len(ck.servers)
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}
