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
	TicksInLast10000         uint32  `json:"ticksInLast10000"`
	EmptyTicksInLast10000    uint32  `json:"emptyTicksInLast10000"`
	Last10000TickQuality     float32 `json:"last10000TickQuality"`
}

type SpectrumData struct {
	CirculatingSupply int64 `json:"circulatingSupply"`
	ActiveAddresses   int   `json:"activeAddresses"`
	Timestamp         int64 `json:"timestamp"`
}

type RichListEntity struct {
	Identity string `bson:"identity" json:"identity"`
	Balance  int64  `bson:"balance" json:"balance"`
}

type RichList []RichListEntity
