package cache

import (
	"sync"
	"time"
)

type Cache struct {
	mutexLock sync.RWMutex

	qubicData           QubicData
	lastQubicDataUpdate time.Time

	spectrumData           SpectrumData
	lastSpectrumDataUpdate time.Time

	richListLength    int32
	richListPageCount int32
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

func (c *Cache) SetRichListLength(richListLength int32) {
	c.mutexLock.Lock()
	defer c.mutexLock.Unlock()

	c.richListLength = richListLength
}

func (c *Cache) GetRichListLength() int32 {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()

	return c.richListLength
}

func (c *Cache) SetRichListPageCount(richListPageCount int32) {
	c.mutexLock.Lock()
	defer c.mutexLock.Unlock()

	c.richListPageCount = richListPageCount
}

func (c *Cache) GetRichListPageCount() int32 {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()

	return c.richListPageCount
}
