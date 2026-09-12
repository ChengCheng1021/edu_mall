package redis

import (
	"fmt"
	"mall/adaptor"
	"mall/config"
	"mall/utils/snowflake"
	"mall/utils/tools"
	"time"

	"github.com/go-redis/redis"
)

type IGenID interface {
	GetNextID() (int64, error)
}

type GenIdNode struct {
	redis    *redis.Client
	idNode   *snowflake.Node
	workerID int
	uuid     string
}

func NewGenIdNode(adaptor adaptor.IAdaptor) *GenIdNode {
	node := &GenIdNode{
		redis: adaptor.GetRedis(),
		uuid:  tools.UUIDHex(),
	}
	err := node.getWorkerID()
	if err != nil {
		panic(err)
	}
	node.idNode, err = snowflake.NewNode(int64(node.workerID))
	if err != nil {
		panic(err)
	}
	return node
}

func fmtGenIdNodeKey(workerID int) string {
	return fmt.Sprintf("%s:genid:%d", config.ServerFullName, workerID)
}

func (g *GenIdNode) getWorkerID() error {
	// 1. 获取到一个没人用的worker_id
	// 2. 保证自己的worker_id是可以被别人知道已经使用过了
	var (
		workerID = 0
	)
	for i := 1; i <= snowflake.WorkerIDMax; i++ {
		redisKey := fmtGenIdNodeKey(i)
		get, _ := g.redis.SetNX(redisKey, g.uuid, time.Second*60).Result()
		if get {
			workerID = i
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	if workerID == 0 {
		return fmt.Errorf("no worker_id")
	}
	g.workerID = workerID
	g.runHeartbeat()
	return nil
}

const (
	ttl       = 60 * time.Second
	renewTick = 30 * time.Second
	maxRetry  = 3
)

func (g *GenIdNode) runHeartbeat() {
	key := fmtGenIdNodeKey(g.workerID)
	go func() {
		tick := time.NewTicker(renewTick)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				for i := 0; i < maxRetry; i++ {
					if err := luaRenew.Run(g.redis, []string{key}, g.uuid, int64(ttl.Seconds())).Err(); err == nil {
						break
					}
					time.Sleep(time.Second * 10)
				}
			}
		}
	}()
}

func (g *GenIdNode) GetNextID() (int64, error) {
	key := fmtGenIdNodeKey(g.workerID)
	ok, err := luaRenew.Run(g.redis, []string{key}, g.uuid, int64(ttl.Seconds())).Result()
	if err != nil || ok.(int64) == 0 {
		return 0, fmt.Errorf("worker id lost, stop generating")
	}
	return g.idNode.GetNextID(), nil
}
