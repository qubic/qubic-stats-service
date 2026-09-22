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
	// Rank is the zero based position in the rich list, highest balance first. Rich lists written
	// before the processor stored it decode as zero, which the api falls back on.
	Rank     int32  `bson:"rank" json:"rank"`
	Identity string `bson:"identity" json:"identity"`
	Balance  int64  `bson:"balance" json:"balance"`
}

type RichList []RichListEntity

// EpochStats is one record of the per epoch supply history the processor keeps. Records are keyed by
// epoch, exist only for epochs that have closed, and never change afterwards.
type EpochStats struct {
	Epoch             uint32 `bson:"_id" json:"epoch"`
	CirculatingSupply int64  `bson:"circulatingSupply" json:"circulatingSupply"`
	TotalEmitted      int64  `bson:"totalEmitted" json:"totalEmitted"`
	ActiveAddresses   int    `bson:"activeAddresses" json:"activeAddresses"`
	EpochEndTimestamp int64  `bson:"epochEndTimestamp" json:"epochEndTimestamp"`
	SpectrumTimestamp int64  `bson:"spectrumTimestamp" json:"spectrumTimestamp"`
	SupplySource      string `bson:"supplySource" json:"supplySource"`
	TimestampSource   string `bson:"timestampSource" json:"timestampSource"`
}

// SupplyHistory is ordered by epoch ascending.
type SupplyHistory []EpochStats
