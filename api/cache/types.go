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
	TicksInLast10000         uint32  `json:"ticksInLast10000"`
	EmptyTicksInLast10000    uint32  `json:"emptyTicksInLast10000"`
	Last10000TickQuality     float32 `json:"last10000TickQuality"`
}

type RichListEntity struct {
	// Rank is the zero based position in the rich list, highest balance first.
	Rank     int32  `bson:"rank" json:"rank"`
	Identity string `bson:"identity" json:"identity"`
	Balance  int64  `bson:"balance" json:"balance"`
}

type RichList []RichListEntity

// EpochStats is one record of the per epoch supply history the processor keeps. Records are keyed by
// epoch and exist only for epochs that have ended.
type EpochStats struct {
	Epoch             uint32 `bson:"_id" json:"epoch"`
	CirculatingSupply int64  `bson:"circulatingSupply" json:"circulatingSupply"`
	TotalEmitted      int64  `bson:"totalEmitted" json:"totalEmitted"`
	ActiveAddresses   int    `bson:"activeAddresses" json:"activeAddresses"`
	EpochEndTimestamp int64  `bson:"epochEndTimestamp" json:"epochEndTimestamp"`
}

// BurnedQUs is the amount of QUs that were emitted up to the end of the epoch but are no longer in
// circulation.
func (e EpochStats) BurnedQUs() int64 {
	return e.TotalEmitted - e.CirculatingSupply
}

// SupplyHistory is ordered by epoch ascending.
type SupplyHistory []EpochStats
