package archiver

type ProcessedTick struct {
	TickNumber uint32 `json:"tickNumber"`
	Epoch      uint32 `json:"epoch"`
}

type SkippedTicksInterval struct {
	StartTick uint32 `json:"startTick"`
	EndTick   uint32 `json:"endTick"`
}

type ProcessedTickIntervals struct {
	InitialProcessedTick uint32 `json:"initialProcessedTick"`
	LastProcessedTick    uint32 `json:"lastProcessedTick"`
}

type ProcessedTickIntervalsPerEpoch struct {
	Epoch     uint32                   `json:"epoch"`
	Intervals []ProcessedTickIntervals `json:"intervals"`
}

type StatusResponse struct {
	LastProcessedTick              *ProcessedTick                    `json:"lastProcessedTick"`
	LastProcessedTicksPerEpoch     map[uint32]uint32                 `json:"lastProcessedTicksPerEpoch"`
	SkippedTicks                   []*SkippedTicksInterval           `json:"skippedTicks"`
	ProcessedTickIntervalsPerEpoch []*ProcessedTickIntervalsPerEpoch `json:"processedTickIntervalsPerEpoch"`
}

type LatestTickResponse struct {
	LatestTick uint32 `json:"latestTick"`
}
