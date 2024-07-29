package cache

import (
	"sync"
	"time"
)

var EmptyPaginationData = RichListPaginationData{}

type RichListPaginationData struct {
	RichListLength    int32
	RichListPageCount int32
}

type Cache struct {
	mutexLock sync.RWMutex

	qubicData           QubicData
	lastQubicDataUpdate time.Time

	spectrumData           SpectrumData
	lastSpectrumDataUpdate time.Time

	richListPaginationDataPerEpoch map[string]RichListPaginationData
}

func (c *Cache) UpdateDataCache(spectrumData SpectrumData, qubicData QubicData) {
	c.mutexLock.Lock()
	defer c.mutexLock.Unlock()

	if spectrumData.Timestamp != 0 {
		c.spectrumData = spectrumData
		c.lastSpectrumDataUpdate = time.Now()
	}
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
func (c *Cache) GetSpectrumData() SpectrumData {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()

	return c.spectrumData
}
func (c *Cache) GetLastQubicDataUpdate() time.Time {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()

	return c.lastQubicDataUpdate

}
func (c *Cache) GetLastSpectrumDataUpdate() time.Time {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()

	return c.lastSpectrumDataUpdate
}

func (c *Cache) SetEpochPaginationData(epoch string, data RichListPaginationData) {
	c.mutexLock.Lock()
	defer c.mutexLock.Unlock()

	c.richListPaginationDataPerEpoch[epoch] = data
}

func (c *Cache) GetEpochPaginationData(epoch string) RichListPaginationData {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()

	return c.richListPaginationDataPerEpoch[epoch]
}
