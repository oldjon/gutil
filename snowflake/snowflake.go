package snowflake

import (
	"fmt"
	"sync"
	"time"
)

const (
	//timeBits     = 41
	nodeBits       = 10              // 1023
	maxStepBits    = 13              // 8191
	NodeMax        = 1<<nodeBits - 1 // 1023
	maxClusterBits = 3
)

var Epoch uint64 = 1684812000000

type Snowflake struct {
	mutex sync.Mutex

	clusterBits  uint64
	clusterMax   uint64
	stepBits     uint64
	stepMask     uint64
	timeShift    uint8
	clusterShift uint8
	nodeShift    uint8

	time    uint64
	cluster uint64
	node    uint64
	step    uint64
}

func New(cluster, node, clusterBits uint64) (*Snowflake, error) {
	if clusterBits > maxClusterBits {
		return nil, fmt.Errorf("cluster bits must be between %d and %d", 0, maxClusterBits)
	}

	clusterMax := uint64(1)<<clusterBits - 1
	if cluster > clusterMax {
		return nil, fmt.Errorf("cluster number must be between %d and %d", 0, clusterMax)
	}

	if node > NodeMax {
		return nil, fmt.Errorf("node number must be between %d and %d", 0, NodeMax)
	}

	// need to make sure maxStepBits >= clusterBits
	// guaranteed for now
	stepBits := maxStepBits - clusterBits
	stepMask := uint64(1)<<stepBits - 1
	timeShift := uint8(clusterBits + nodeBits + stepBits)
	clusterShift := uint8(nodeBits + stepBits)
	nodeShift := uint8(stepBits)

	return &Snowflake{
		clusterBits:  clusterBits,
		clusterMax:   clusterMax,
		stepBits:     stepBits,
		stepMask:     stepMask,
		timeShift:    timeShift,
		clusterShift: clusterShift,
		nodeShift:    nodeShift,

		time:    0,
		cluster: cluster,
		node:    node,
		step:    0,
	}, nil
}

func (sf *Snowflake) Next() uint64 {
	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	now := ms()
	switch {
	case now < sf.time:
		now = wait(now, sf.time)
	case now == sf.time:
		sf.step = (sf.step + 1) & sf.stepMask
		if sf.step == 0 {
			now = wait(now, sf.time)
		}
	case now > sf.time:
		sf.step = 1
	}
	sf.time = now

	id := (now-Epoch)<<sf.timeShift |
		sf.cluster<<sf.clusterShift |
		sf.node<<sf.nodeShift |
		sf.step

	return id
}

func ms() uint64 {
	return uint64(time.Now().UnixMilli())
}

func wait(now uint64, target uint64) uint64 {
	for now <= target {
		now = ms()
	}
	return now
}
