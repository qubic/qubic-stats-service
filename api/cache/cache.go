package cache

import (
	"sync"
	"time"
)

type Cache struct {
	mutexLock sync.RWMutex

	qubicData           QubicData
	lastQubicDataUpdate time.Time

	supplyHistory           SupplyHistory
	lastSupplyHistoryUpdate time.Time
}

func (c *Cache) UpdateQubicData(qubicData QubicData) {
	c.mutexLock.Lock()
	defer c.mutexLock.Unlock()

	if qubicData.Timestamp != 0 {
		c.qubicData = qubicData
		c.lastQubicDataUpdate = time.Now()
	}
}
func (c *Cache) GetQubicData() QubicData {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()

	return c.qubicData
}
func (c *Cache) GetLastQubicDataUpdate() time.Time {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()

	return c.lastQubicDataUpdate

}

// UpdateSupplyHistory replaces the cached supply history. The slice is never modified in place, so
// readers may hold on to the one they were handed.
func (c *Cache) UpdateSupplyHistory(supplyHistory SupplyHistory) {
	c.mutexLock.Lock()
	defer c.mutexLock.Unlock()

	c.supplyHistory = supplyHistory
	c.lastSupplyHistoryUpdate = time.Now()
}

func (c *Cache) GetSupplyHistory() SupplyHistory {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()

	return c.supplyHistory
}

// GetLatestEpochStats returns the record of the most recent completed epoch, which the circulating
// supply, the active addresses and the rich list are based on. It reports false when there is none.
func (c *Cache) GetLatestEpochStats() (EpochStats, bool) {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()

	if len(c.supplyHistory) == 0 {
		return EpochStats{}, false
	}
	return c.supplyHistory[len(c.supplyHistory)-1], true
}

func (c *Cache) GetLastSupplyHistoryUpdate() time.Time {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()

	return c.lastSupplyHistoryUpdate
}
