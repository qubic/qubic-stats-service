package rpc

import (
	"github.com/qubic/qubic-stats-api/types"
	"sync"
	"time"
)

type DataCache struct {
	mutexLock sync.RWMutex

	qubicData           *types.QubicData
	lastQubicDataUpdate time.Time

	spectrumData           *types.SpectrumData
	lastSpectrumDataUpdate time.Time
}

func (c *DataCache) UpdateDataCache(spectrumData *types.SpectrumData, qubicData *types.QubicData) {

	c.mutexLock.Lock()
	defer c.mutexLock.Unlock()

	if spectrumData != nil {
		c.spectrumData = spectrumData
		c.lastSpectrumDataUpdate = time.Now()
	}
	if qubicData != nil {
		c.qubicData = qubicData
		c.lastQubicDataUpdate = time.Now()
	}

}

func (c *DataCache) getQubicData() *types.QubicData {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()
	return c.qubicData

}
func (c *DataCache) getSpectrumData() *types.SpectrumData {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()
	return c.spectrumData

}
func (c *DataCache) getLastQubicDataUpdate() time.Time {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()
	return c.lastQubicDataUpdate

}
func (c *DataCache) getLastSpectrumDataUpdate() time.Time {
	c.mutexLock.RLock()
	defer c.mutexLock.RUnlock()
	return c.lastSpectrumDataUpdate

}
