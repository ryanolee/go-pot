package stall

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Structured map for stallers mapped by identifierAddress and Connection ID
type StallerCollection struct {
	stallers map[string]map[uint64]Staller
	// The maximum number of stallers allowed per group
	groupLimit int
	lock       sync.Mutex
}

func NewStallerCollection(groupLimit int) *StallerCollection {
	return &StallerCollection{
		groupLimit: groupLimit,
		stallers:   make(map[string]map[uint64]Staller),
	}
}

func (c *StallerCollection) Add(staller Staller) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.stallers[staller.GetGroupIdentifier()]; !ok {
		c.stallers[staller.GetGroupIdentifier()] = make(map[uint64]Staller)
	}

	if len(c.stallers[staller.GetGroupIdentifier()]) > c.groupLimit {
		return fmt.Errorf("failed to add id %d to group %s due to group limit %d being reached", staller.GetIdentifier(), staller.GetGroupIdentifier(), c.groupLimit)
	}

	c.stallers[staller.GetGroupIdentifier()][staller.GetIdentifier()] = staller
	return nil
}

func (c *StallerCollection) Delete(staller Staller) {
	zap.L().Sugar().Debugw("Deleting staller", "groupId", staller.GetIdentifier(), "id", staller.GetIdentifier())
	c.lock.Lock()
	defer c.lock.Unlock()

	if identifierMap, ok := c.stallers[staller.GetGroupIdentifier()]; ok {
		delete(identifierMap, staller.GetIdentifier())
	}

	if len(c.stallers[staller.GetGroupIdentifier()]) == 0 {
		delete(c.stallers, staller.GetGroupIdentifier())
	}
}

func (c *StallerCollection) PruneNByIdentifier(count int) {
	for i := 0; i < count; i++ {
		c.PruneByIdentifier()
	}
}

func (c *StallerCollection) Len() int {
	count := 0
	for _, identifierMap := range c.stallers {
		count += len(identifierMap)
	}
	return count
}

func (c *StallerCollection) PruneByIdentifierGroup(id string) {
	stallers, ok := c.stallers[id]
	if !ok {
		return
	}

	for _, staller := range stallers {
		c.Delete(staller)
	}
}

func (c *StallerCollection) PruneByIdentifier() {
	stallers := c.getMostActiveStallers()

	for _, staller := range stallers {
		staller.Close()
		c.Delete(staller)
		break
	}
}

func (c *StallerCollection) getMostActiveStallers() map[uint64]Staller {
	var mostActiveStallers map[uint64]Staller
	for _, identifierMap := range c.stallers {
		if len(identifierMap) > len(mostActiveStallers) {
			mostActiveStallers = identifierMap
		}
	}

	return mostActiveStallers
}
