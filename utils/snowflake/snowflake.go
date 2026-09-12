package snowflake

import (
	"errors"
	"sync"
	"time"
)

const (
	// 自定义纪元，设置为 2024-01-01 00:00:00.000 的毫秒时间戳，以减小时间戳部分的位数
	epoch        int64 = 1704038400000
	workerBits   uint8 = 5 // 5 位 workerID，支持 32 个节点
	sequenceBits uint8 = 9 // 9 位序列号，单毫秒内支持 512 个 ID（满足每秒 100 个的需求）

	WorkerIDMax  = -1 ^ (-1 << workerBits)   // 31
	sequenceMask = -1 ^ (-1 << sequenceBits) // 511
)

type Node struct {
	mu        sync.Mutex
	workerID  int64
	lastStamp int64
	sequence  int64
}

// NewNode 传入 workerID（0~31）即可
func NewNode(workerID int64) (*Node, error) {
	if workerID < 0 || workerID > WorkerIDMax {
		return nil, errors.New("worker id out of range")
	}
	return &Node{workerID: workerID}, nil
}

// GetNextID 生成下一个 ID (总长度控制在 14 位以内)
func (n *Node) GetNextID() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < n.lastStamp {
		time.Sleep(time.Duration(n.lastStamp-now) * time.Millisecond)
		now = time.Now().UnixMilli()
	}

	if now == n.lastStamp {
		n.sequence = (n.sequence + 1) & sequenceMask
		if n.sequence == 0 {
			for now <= n.lastStamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		n.sequence = 0
	}

	n.lastStamp = now
	// 位移计算：(时间戳差值) << (workerBits + sequenceBits) | (workerID) << sequenceBits | sequence
	return (now-epoch)<<(workerBits+sequenceBits) | n.workerID<<sequenceBits | n.sequence
}
