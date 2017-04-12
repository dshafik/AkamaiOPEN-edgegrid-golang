package edgegrid

type GtmResource struct {
	Name                        string                  `json:"name"`
	AggregationType             GtmAggregationTypeValue `json:"aggregationType,omitempty"`
	ConstrainedProperty         string                  `json:"constrainedProperty,omitempty"`
	DecayRate                   float64                 `json:"decayRate,omitempty"`   // Value between 0 and 1
	Description                 string                  `json:"description,omitempty"` // Maximum 256 chars
	HostHeader                  string                  `json:"hostHeader,omitempty"`  // Maximum 255 chars
	LeaderString                string                  `json:"leaderString,omitempty"`
	LeastSquaresDecay           float64                 `json:"leastSquaresDecay,omitempty"`
	LoadImbalancePercentage     int                     `json:"loadImbalancePercentage,omitempty"`
	MaxUMultiplicativeIncrement int                     `json:"maxUMultiplicativeIncrement,omitempty"`
	ResourceInstances           []*GtmResourceInstance  `json:"resourceInstances,omitempty"`
	Type                        GtmResourceTypeValue    `json:"type,omitempty"`
	UpperBound                  int                     `json:"upperBound,omitempty"`
	Links                       []*GtmHypermediaLinks   `json:"links,omitempty"`
}

type GtmResourceInstance struct {
	DatacenterID         int      `json:"datacenterId"`
	LoadObject           string   `json:"loadObject"`
	LoadObjectPort       int      `json:"loadObjectPort,omitempty"`
	LoadServers          []string `json:"loadServers,omitempty"`
	UseDefaultLoadObject bool     `json:"useDefaultLoadObject,omitempty"`
}

type GtmAggregationTypeValue string
type GtmResourceTypeValue string

const (
	GTM_AGGREGATION_TYPE_SUM    GtmAggregationTypeValue = "sum"
	GTM_AGGREGATION_TYPE_MEDIAN GtmAggregationTypeValue = "median"
	GTM_AGGREGATION_TYPE_LATEST GtmAggregationTypeValue = "latest"

	GTM_RESOURCE_TYPE_XML_VIA_HTTP      GtmResourceTypeValue = "XML load object via HTTP"
	GTM_RESOURCE_TYPE_XML_VIA_HTTPS     GtmResourceTypeValue = "XML load object via HTTPS"
	GTM_RESOURCE_TYPE_NON_XML_VIA_HTTP  GtmResourceTypeValue = "Non-XML load object via HTTP"
	GTM_RESOURCE_TYPE_NON_XML_VIA_HTTPS GtmResourceTypeValue = "Non-XML load object via HTTPS"
	GTM_RESOURCE_TYPE_DOWNLOAD_SCORE    GtmResourceTypeValue = "Download score"
)
