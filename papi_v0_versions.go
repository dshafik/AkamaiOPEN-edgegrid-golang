package edgegrid

import "time"

type PapiVersions struct {
	service      *PapiV0Service
	PropertyId   string `json:"propertyId"`
	PropertyName string `json:"propertyName"`
	AccountId    string `json:"accountId"`
	ContractId   string `json:"contractId"`
	GroupId      string `json:"groupId"`
	Versions     struct {
		Items []*PapiVersion `json:"items"`
	} `json:"versions"`
	RuleFormat string `json"ruleFormat,omitempty"`
}

type PapiVersion struct {
	parent                *PapiVersions
	PropertyVersion       int       `json:"propertyVersion,omitempty"`
	UpdatedByUser         string    `json:"updatedByUser,omitempty"`
	UpdatedDate           time.Time `json:"updatedDate,omitempty"`
	ProductionStatus      string    `json:"productionStatus,omitempty"`
	StagingStatus         string    `json:"stagingStatus,omitempty"`
	Etag                  string    `json:"etag,omitempty"`
	ProductId             string    `json:"productId,omitempty"`
	Note                  string    `json:"note,omitempty"`
	CreateFromVersion     int       `json:"createFromVersion,omitempty"`
	CreateFromVersionEtag string    `json:"createFromVersionEtag,omitempty"`
}
