package edgegrid

import "encoding/json"

type PapiEdgeHostnames struct {
	service       *PapiV0Service
	AccountId     string `json:"accountId"`
	ContractId    string `json:"contractId"`
	GroupId       string `json:"groupId"`
	EdgeHostnames struct {
		Items []*PapiEdgeHostname `json:"items"`
	} `json:"edgeHostnames"`
}

func (edgeHostnames *PapiEdgeHostnames) UnmarshalJSON(b []byte) error {
	type PapiEdgeHostnamesTemp PapiEdgeHostnames
	temp := &PapiEdgeHostnamesTemp{service: edgeHostnames.service}

	if err := json.Unmarshal(b, temp); err != nil {
		return err
	}
	*edgeHostnames = (PapiEdgeHostnames)(*temp)

	for key, _ := range edgeHostnames.EdgeHostnames.Items {
		edgeHostnames.EdgeHostnames.Items[key].parent = edgeHostnames

	}

	return nil
}

func (edgeHostnames *PapiEdgeHostnames) NewEdgeHostname() (*PapiEdgeHostname, error) {
	edgeHostname := &PapiEdgeHostname{parent: edgeHostnames}
	return edgeHostname, nil
}

type PapiEdgeHostname struct {
	parent             *PapiEdgeHostnames
	EdgeHostnameId     string `json:"edgeHostnameId,omitempty"`
	EdgeHostnameDomain string `json:"edgeHostnameDomain,omitempty"`
	ProductId          string `json:"productId"`
	DomainPrefix       string `json:"domainPrefix"`
	DomainSuffix       string `json:"domainSuffix"`
	Status             string `json:"status,omitempty"`
	Secure             bool   `json:"secure,omitempty"`
	IPVersionBehavior  string `json:"ipVersionBehavior,omitempty"`
}
