package edgegrid

type GtmProperties struct {
	service *GtmV11Service
	Items   []*GtmProperty `json:"items"`
}

type GtmProperty struct {
	parent                    *GtmV11Service
	HandoutMode               GtmHandoutModeValue          `json:"handoutMode"`
	Name                      string                       `json:"name"`
	ScoreAggregationType      GtmScoreAggregationTypeValue `json:"scoreAggregationType"`
	TrafficTargets            []*GtmTrafficTarget          `json:"trafficTargets"`
	Type                      GtmPropertyTypeValue         `json:"type"`
	BackupCName               string                       `json:"backupCName,omitempty"`
	BackupIP                  string                       `json:"backupIp,omitempty"`
	BalanceByDownloadScore    bool                         `json:"balanceByDownloadScore,omitempty"`
	Cname                     string                       `json:"cname,omitempty"`
	Comments                  string                       `json:"comments,omitempty"`
	DynamicTTL                int                          `json:"dynamicTTL,omitempty"`
	FailbackDelay             int                          `json:"failbackDelay,omitempty"`
	FailoverDelay             int                          `json:"failoverDelay,omitempty"`
	HealthMax                 int                          `json:"healthMax,omitempty"`
	HealthMultiplier          int                          `json:"healthMultiplier,omitempty"`
	HealthThreshold           int                          `json:"healthThreshold,omitempty"`
	Ipv6                      bool                         `json:"ipv6,omitempty"`
	LastModified              string                       `json:"lastModified,omitempty"`
	LivenessTests             []*GtmLivenessTest           `json:"livenessTests,omitempty"`
	LoadImbalancePercentage   int                          `json:"loadImbalancePercentage,omitempty"`
	MapName                   string                       `json:"mapName,omitempty"`
	MaxUnreachablePenalty     int                          `json:"maxUnreachablePenalty,omitempty"`
	MxRecords                 []*GtmMxRecord               `json:"mxRecords,omitempty"`
	StaticTTL                 int                          `json:"staticTTL,omitempty"`
	StickinessBonusConstant   int                          `json:"stickinessBonusConstant,omitempty"`
	StickinessBonusPercentage int                          `json:"stickinessBonusPercentage,omitempty"`
	UnreachableThreshold      int                          `json:"unreachableThreshold,omitempty"`
	UseComputedTargets        bool                         `json:"useComputedTargets,omitempty"`
}

type GtmHandoutModeValue string
type GtmScoreAggregationTypeValue string
type GtmPropertyTypeValue string

const (
	GTM_HANDOUT_MODE_NORMAL        GtmHandoutModeValue = "normal"
	GTM_HANDOUT_MODE_PERSISTENT    GtmHandoutModeValue = "persistent"
	GTM_HANDOUT_MODE_ONE_IP        GtmHandoutModeValue = "one-ip"
	GTM_HANDOUT_MODE_ONE_IP_HASHED GtmHandoutModeValue = "one-ip-hashed"
	GTM_HANDOUT_MODE_ALL_LIVE_IPS  GtmHandoutModeValue = "all-live-ips"

	GTM_SCORE_AGGREGRATION_TYPE_MEAN   GtmScoreAggregationTypeValue = "mean"
	GTM_SCORE_AGGREGRATION_TYPE_MEDIAN GtmScoreAggregationTypeValue = "median"
	GTM_SCORE_AGGREGRATION_TYPE_BEST   GtmScoreAggregationTypeValue = "best"
	GTM_SCORE_AGGREGRATION_TYPE_WORST  GtmScoreAggregationTypeValue = "worst"

	GTM_PROPERTY_TYPE_FAILOVER                           GtmPropertyTypeValue = "failover"
	GTM_PROPERTY_TYPE_GEOGRAPHIC                         GtmPropertyTypeValue = "geographic"
	GTM_PROPERTY_TYPE_CIDRMAPPING                        GtmPropertyTypeValue = "cidrmapping"
	GTM_PROPERTY_TYPE_WEIGHTED_ROUND_ROBIN               GtmPropertyTypeValue = "weighted-round-robin"
	GTM_PROPERTY_TYPE_WEIGHTED_HASHED                    GtmPropertyTypeValue = "weighted-hashed"
	GTM_PROPERTY_TYPE_WEIGHTED_ROUND_ROBIN_LOAD_FEEDBACK GtmPropertyTypeValue = "weighted-round-robin-load-feedback"
	GTM_PROPERTY_TYPE_QTR                                GtmPropertyTypeValue = "qtr"
	GTM_PROPERTY_TYPE_PERFORMANCE                        GtmPropertyTypeValue = "performance"
	GTM_PROPERTY_TYPE_ASMAPPING                          GtmPropertyTypeValue = "asmapping"
)
