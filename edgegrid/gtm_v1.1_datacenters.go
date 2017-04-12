package edgegrid

type GtmDatacenters struct {
	service *GtmV11Service
	Domain  string           `json:"domain"`
	Items   []*GtmDatacenter `json:"items",omitempty`
}

type GtmDatacenter struct {
	parent               *GtmDatacenters
	DatacenterId         int                   `json:datacenterId`
	City                 string                `json:"city,omitempty"`
	CloneOf              int                   `json:"cloneOf,omitempty"`
	CloudServerTargeting bool                  `json:"cloudServerTargeting,omitempty"`
	Continent            string                `json:"continent,omitempty"`
	Country              string                `json:"country,omitempty"`
	DefaultLoadObject    *GtmDefaultLoadObject `json:"defaultLoadObject,omitempty"`
	Latitude             float64               `json:"latitude,omitempty"`
	Longitude            float64               `json:"longitude,omitempty"`
	Nickname             string                `json:"nickname,omitempty"`
	StateOrProvince      string                `json:"stateOrProvince,omitempty"`
	Virtual              bool                  `json:"virtual,omitempty"`
}

type GtmDefaultLoadObject struct {
	LoadObject     string   `json:"loadObject"`
	LoadObjectPort int      `json:"loadObjectPort"`
	LoadServers    []string `json:"loadServers"`
}

type GtmDatacenterDefault struct {
	GtmDatacenter
	Nickname string `json:"nickname"`
}
