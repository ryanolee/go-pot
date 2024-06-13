package stall

import (
	"sync"

	"go.uber.org/zap"
)

// Structured map for stallers mapped by ipAddress and Connection ID
type StallerCollection struct {
	stallers map[string]map[uint64]Staller
	lock     sync.Mutex
}

func NewStallerCollection() *StallerCollection {
	return &StallerCollection{
		stallers: make(map[string]map[uint64]Staller),
	}
}

func (c *StallerCollection) Add(staller Staller) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.stallers[staller.GetGroupIdentifier()]; !ok {
		c.stallers[staller.GetGroupIdentifier()] = make(map[uint64]Staller)
	}

	c.stallers[staller.GetGroupIdentifier()][staller.GetIdentifier()] = staller
}

func (c *StallerCollection) Delete(staller Staller) {
	zap.L().Sugar().Debugw("Deleting staller", "groupId", staller.GetIdentifier(), "id", staller.GetIdentifier())
	c.lock.Lock()
	defer c.lock.Unlock()

	if ipMap, ok := c.stallers[staller.GetGroupIdentifier()]; ok {
		delete(ipMap, staller.GetIdentifier())
	}

	if len(c.stallers[staller.GetGroupIdentifier()]) == 0 {
		delete(c.stallers, staller.GetGroupIdentifier())
	}
}

func (c *StallerCollection) PruneNByIp(count int) {
	for i := 0; i < count; i++ {
		c.PruneByIp()
	}
}

func (c *StallerCollection) Len() int {
	count := 0
	for _, ipMap := range c.stallers {
		count += len(ipMap)
	}
	return count
}

func (c *StallerCollection) PruneByIp() {
	stallers := c.getMostActiveStallers()

	for _, staller := range stallers {
		staller.Close()
		c.Delete(staller)
		break
	}
}

func (c *StallerCollection) getMostActiveStallers() map[uint64]Staller {
	var mostActiveStallers map[uint64]Staller
	for _, ipMap := range c.stallers {
		if len(ipMap) > len(mostActiveStallers) {
			mostActiveStallers = ipMap
		}
	}

	return mostActiveStallers
}
