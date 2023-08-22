package snowflake

import (
	"fmt"
	"sync"
)

const (
	timeBits53       = 39 // 16 years
	timeMask53       = 1<<timeBits53 - 1
	nodeBits53       = 5                 // 32
	NodeMax53        = 1<<nodeBits53 - 1 // 1023
	maxStepBits53    = 9                 // 512
	maxClusterBits53 = 3
)

type Snowflake53 struct {
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

func New53(cluster, node, clusterBits uint64) (*Snowflake53, error) {
	if clusterBits > maxClusterBits53 {
		return nil, fmt.Errorf("cluster bits must be between %d and %d", 0, maxClusterBits53)
	}

	clusterMax := uint64(1)<<clusterBits - 1
	if cluster > clusterMax {
		return nil, fmt.Errorf("cluster number must be between %d and %d", 0, clusterMax)
	}

	if node > NodeMax53 {
		return nil, fmt.Errorf("node number must be between %d and %d", 0, NodeMax53)
	}

	// need to make sure maxStepBits >= clusterBits
	// guaranteed for now
	stepBits := maxStepBits53 - clusterBits
	stepMask := uint64(1)<<stepBits - 1
	timeShift := uint8(clusterBits + nodeBits53 + stepBits)
	clusterShift := uint8(nodeBits53 + stepBits)
	nodeShift := uint8(stepBits)

	return &Snowflake53{
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

func (sf *Snowflake53) Next() uint64 {
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

	id := ((now-Epoch)&timeMask53)<<sf.timeShift |
		sf.cluster<<sf.clusterShift |
		sf.node<<sf.nodeShift |
		sf.step

	return id
}
