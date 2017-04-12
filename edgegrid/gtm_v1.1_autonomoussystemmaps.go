package edgegrid

type GtmAutonomousSystemMaps struct {
	service           *GtmV11Service
	Assignments       []*GtmAutonomousSystemAssignment `json:"assignments"`
	DefaultDatacenter *GtmDatacenterDefault            `json:"defaultDatacenter"`
	Name              string                           `json:"name"`
}

type GtmAutonomousSystemAssignment struct {
	AsNumbers    int    `json:"asNumbers"` // Range: 1 to 4294967295
	DatacenterId int    `json:"datacenterId"`
	nickname     string `json:"nickname"`
}
