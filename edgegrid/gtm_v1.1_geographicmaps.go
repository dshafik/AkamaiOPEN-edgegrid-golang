package edgegrid

type GtmGeographicMaps struct {
	service *GtmV11Service
	Items   []*GtmGeographicMap `json:"items"`
}

type GtmGeographicMap struct {
	parent            *GtmGeographicMaps
	Assignments       []*GtmGeographicMapAssignment `json:"assignments"`
	DefaultDatacenter GtmDatacenterDefault          `json:"defaultDatacenter"`
	Links             []*GtmHypermediaLinks         `json:"links"`
	Name              string                        `json:"name"`
}

type GtmGeographicMapAssignment struct {
	Countries    []string `json:"countries"`
	DatacenterID int      `json:"datacenterId"`
	Nickname     string   `json:"nickname"`
}
