package edgegrid

type PapiHostnames struct {
	AccountId       string `json:"accountId"`
	ContractId      string `json:"contractId"`
	GroupId         string `json:"groupId"`
	PropertyId      string `json:"propertyId"`
	PropertyVersion int    `json:"propertyVersion"`
	Etag            string `json:"etag"`
	Hostnames       struct {
		Items []*PapiHostname `json:"items"`
	} `json:"hostnames"`
}

type PapiHostname struct {
	CnameType      string `json:"cnameType"`
	EdgeHostnameId string `json:"edgeHostnameId"`
	CnameFrom      string `json:"cnameFrom"`
	CnameTo        string `json:"cnameTo,omitempty"`
}
