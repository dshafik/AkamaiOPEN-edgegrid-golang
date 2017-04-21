package edgegrid

import (
	"fmt"

	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

type PapiEdgeHostnames struct {
	Resource
	service       *PapiV0Service
	AccountID     string `json:"accountId"`
	ContractID    string `json:"contractId"`
	GroupID       string `json:"groupId"`
	EdgeHostnames struct {
		Items []*PapiEdgeHostname `json:"items"`
	} `json:"edgeHostnames"`
}

func NewPapiEdgeHostnames(service *PapiV0Service) *PapiEdgeHostnames {
	edgeHostnames := &PapiEdgeHostnames{service: service}
	edgeHostnames.Init()
	return edgeHostnames
}

func (edgeHostnames *PapiEdgeHostnames) PostUnmarshalJSON() error {
	edgeHostnames.Init()

	for key, edgeHostname := range edgeHostnames.EdgeHostnames.Items {
		edgeHostnames.EdgeHostnames.Items[key].parent = edgeHostnames

		if edgeHostname, ok := json.ImplementsPostJSONUnmarshaler(edgeHostname); ok {
			if err := edgeHostname.(json.PostJSONUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}

	edgeHostnames.Complete <- true

	return nil
}

func (edgeHostnames *PapiEdgeHostnames) GetEdgeHostnames(contract *PapiContract, group *PapiGroup, options string) error {
	if contract == nil {
		contract = NewPapiContract(NewPapiContracts(edgeHostnames.service))
		contract.ContractID = group.ContractIDs[0]
	}

	if options != "" {
		options = fmt.Sprintf("&options=%s", options)
	}

	res, err := edgeHostnames.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/edgehostnames?groupId=%s&contractId=%s%s",
			group.GroupID,
			contract.ContractID,
			options,
		),
	)
	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newEdgeHostnames := NewPapiEdgeHostnames(edgeHostnames.service)
	err = res.BodyJSON(newEdgeHostnames)
	if err != nil {
		return err
	}

	*edgeHostnames = *newEdgeHostnames

	return nil
}

type PapiEdgeHostname struct {
	Resource
	parent             *PapiEdgeHostnames
	EdgeHostnameID     string `json:"edgeHostnameId,omitempty"`
	EdgeHostnameDomain string `json:"edgeHostnameDomain,omitempty"`
	ProductID          string `json:"productId"`
	DomainPrefix       string `json:"domainPrefix"`
	DomainSuffix       string `json:"domainSuffix"`
	Status             string `json:"status,omitempty"`
	Secure             bool   `json:"secure,omitempty"`
	IPVersionBehavior  string `json:"ipVersionBehavior,omitempty"`
}

func NewPapiEdgeHostname(edgeHostnames *PapiEdgeHostnames) (*PapiEdgeHostname, error) {
	edgeHostname := &PapiEdgeHostname{parent: edgeHostnames}
	edgeHostname.Init()
	return edgeHostname, nil
}
