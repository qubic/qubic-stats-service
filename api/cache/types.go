package cache

type QubicData struct {
	Timestamp                int64   `json:"timestamp"`
	Price                    float32 `json:"price"`
	MarketCap                int64   `json:"marketCap"`
	Epoch                    uint32  `json:"epoch"`
	CurrentTick              uint32  `json:"currentTick"`
	TicksInCurrentEpoch      uint32  `json:"ticksInCurrentEpoch"`
	EmptyTicksInCurrentEpoch uint32  `json:"emptyTicksInCurrentEpoch"`
	EpochTickQuality         float32 `json:"epochTickQuality"`
	BurnedQUs                uint64  `json:"burnedQUs"`
}

type SpectrumData struct {
	CirculatingSupply int64 `json:"circulatingSupply"`
	ActiveAddresses   int   `json:"activeAddresses"`
	Timestamp         int64 `json:"timestamp"`
}
