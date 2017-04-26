package edgegrid

import (
	"fmt"

	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

// PapiEdgeHostnames is a collection for PAPI Edge Hostname resources
type PapiEdgeHostnames struct {
	resource
	service       *PapiV0Service
	AccountID     string `json:"accountId"`
	ContractID    string `json:"contractId"`
	GroupID       string `json:"groupId"`
	EdgeHostnames struct {
		Items []*PapiEdgeHostname `json:"items"`
	} `json:"edgeHostnames"`
}

// NewPapiEdgeHostnames creates a new PapiEdgeHostnames
func NewPapiEdgeHostnames(service *PapiV0Service) *PapiEdgeHostnames {
	edgeHostnames := &PapiEdgeHostnames{service: service}
	edgeHostnames.Init()
	return edgeHostnames
}

// PostUnmarshalJSON is called after JSON unmarshaling into PapiEdgeHostnames
//
// See: edgegrid/json.Unmarshal()
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

// NewEdgeHostname creates a new PapiEdgeHostname within a given PapiEdgeHostnames
func (edgeHostnames *PapiEdgeHostnames) NewEdgeHostname() *PapiEdgeHostname {
	edgeHostname := NewPapiEdgeHostname(edgeHostnames)
	edgeHostnames.EdgeHostnames.Items = append(edgeHostnames.EdgeHostnames.Items, edgeHostname)
	return edgeHostname
}

// GetEdgeHostnames will populate PapiEdgeHostnames with Edge Hostname data
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listedgehostnames
// Endpoint: GET /papi/v0/edgehostnames/{?contractId,groupId,options}
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
	if err = res.BodyJSON(newEdgeHostnames); err != nil {
		return err
	}

	*edgeHostnames = *newEdgeHostnames

	return nil
}

// PapiEdgeHostname represents an Edge Hostname resource
type PapiEdgeHostname struct {
	resource
	parent                 *PapiEdgeHostnames
	EdgeHostnameID         string `json:"edgeHostnameId,omitempty"`
	EdgeHostnameDomain     string `json:"edgeHostnameDomain,omitempty"`
	ProductID              string `json:"productId"`
	DomainPrefix           string `json:"domainPrefix"`
	DomainSuffix           string `json:"domainSuffix"`
	Status                 string `json:"status,omitempty"`
	Secure                 bool   `json:"secure,omitempty"`
	IPVersionBehavior      string `json:"ipVersionBehavior,omitempty"`
	MapDetailsSerialNumber int    `json:"mapDetails:serialNumber,omitempty"`
	MapDetailsSlotNumber   int    `json:"mapDetails:slotNumber,omitempty"`
	MapDetailsMapDomain    string `json:"mapDetails:mapDomain,omitempty"`
}

// NewPapiEdgeHostname creates a new PapiEdgeHostname
func NewPapiEdgeHostname(edgeHostnames *PapiEdgeHostnames) *PapiEdgeHostname {
	edgeHostname := &PapiEdgeHostname{parent: edgeHostnames}
	edgeHostname.Init()
	return edgeHostname
}

// GetEdgeHostname populates PapiEdgeHostname with data
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getanedgehostname
// Endpoint: GET /papi/v0/edgehostnames/{edgeHostnameId}{?contractId,groupId,options}
func (edgeHostname *PapiEdgeHostname) GetEdgeHostname(options string) error {
	if options != "" {
		options = "&options=" + options
	}

	res, err := edgeHostname.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/edgehostnames/%s?contractId=%s&groupId=%s%s",
			edgeHostname.EdgeHostnameID,
			edgeHostname.parent.ContractID,
			edgeHostname.parent.GroupID,
			options,
		),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newEdgeHostname := NewPapiEdgeHostname(edgeHostname.parent)
	if err := res.BodyJSON(newEdgeHostname); err != nil {
		return err
	}

	*edgeHostname = *newEdgeHostname

	return nil
}

// Save creates a new Edge Hostname
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#createanewedgehostname
// Endpoint: POST /papi/v0/edgehostnames/{?contractId,groupId,options}
func (edgeHostname *PapiEdgeHostname) Save(options string) error {
	if options != "" {
		options = "&options=" + options
	}
	res, err := edgeHostname.parent.service.client.PostJSON(
		fmt.Sprintf(
			"/papi/v0/edgehostnames/?contractId=%s&groupId=%s%s",
			edgeHostname.parent.ContractID,
			edgeHostname.parent.GroupID,
			options,
		),
		edgeHostname,
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	var location JSONBody
	if err = res.BodyJSON(&location); err != nil {
		return err
	}

	res, err = edgeHostname.parent.service.client.Get(
		location["edgeHostnameLink"].(string),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	edgeHostnames := NewPapiEdgeHostnames(edgeHostname.parent.service)
	if err = res.BodyJSON(edgeHostnames); err != nil {
		return err
	}

	newEdgehostname := edgeHostnames.EdgeHostnames.Items[0]
	newEdgehostname.parent = edgeHostname.parent
	edgeHostname.parent.EdgeHostnames.Items = append(edgeHostname.parent.EdgeHostnames.Items, edgeHostnames.EdgeHostnames.Items...)

	*edgeHostname = *newEdgehostname

	return nil
}
