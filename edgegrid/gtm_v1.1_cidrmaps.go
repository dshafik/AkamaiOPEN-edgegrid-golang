package edgegrid

type GtmCidrMaps struct {
	service *GtmV11Service
	Items   []*GtmCidrMap `json:"items"`
}

type GtmCidrMap struct {
	parent            *GtmCidrMap
	Assignments       []*GtmCidrAssignment  `json:"assignments"`
	DefaultDatacenter *GtmDatacenterDefault `json:"defaultDatacenter"`
	Name              string                `json:"name"`
	Links             []*GtmHypermediaLinks `json:"links,omitempty"`
}

type GtmCidrAssignment struct {
	Blocks       []string `json:"blocks"`
	DatacenterID int      `json:"datacenterId"`
	Nickname     string   `json:"nickname"`
}
