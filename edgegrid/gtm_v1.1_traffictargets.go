package edgegrid

type GtmTrafficTarget struct {
	DatacenterID int      `json:"datacenterId"`
	Enabled      bool     `json:"enabled"`
	HandoutCName string   `json:"handoutCName"`
	Name         string   `json:"name"`
	Servers      []string `json:"servers"`
	Weight       float64  `json:"weight"`
}
