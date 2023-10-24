package http

import (
	"runtime/debug"
	"sync"

	"go.uber.org/zap"
)

// Structured map for stallers mapped by ipAddress and Connection ID
type StallerCollection struct {
	stallers map[string]map[uint64]*HttpStaller
	lock     sync.Mutex
}

func NewStallerCollection() *StallerCollection {
	return &StallerCollection{
		stallers: make(map[string]map[uint64]*HttpStaller),
	}
}

func (c *StallerCollection) Add(staller *HttpStaller) {
	zap.L().Sugar().Infow("Adding staller", "ipAddress", staller.ipAddress, "id", staller.id)
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.stallers[staller.ipAddress]; !ok {
		c.stallers[staller.ipAddress] = make(map[uint64]*HttpStaller)
	}

	c.stallers[staller.ipAddress][staller.id] = staller
}

func (c *StallerCollection) Delete(staller *HttpStaller) {
	zap.L().Sugar().Infow("Deleting staller", "ipAddress", staller.ipAddress, "id", staller.id)
	c.lock.Lock()
	defer c.lock.Unlock()

	if ipMap, ok := c.stallers[staller.ipAddress]; ok {
		delete(ipMap, staller.id)
	}

	if len(c.stallers[staller.ipAddress]) == 0 {
		delete(c.stallers, staller.ipAddress)
	}
}

func (c *StallerCollection) PruneNByIp(count int) {
	for i := 0; i < count; i++ {
		c.PruneByIp()
	}

	debug.FreeOSMemory()
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

func (c *StallerCollection) getMostActiveStallers() map[uint64]*HttpStaller {
	var mostActiveStallers map[uint64]*HttpStaller
	for _, ipMap := range c.stallers {
		if len(ipMap) > len(mostActiveStallers) {
			mostActiveStallers = ipMap
		}
	}

	return mostActiveStallers
}
