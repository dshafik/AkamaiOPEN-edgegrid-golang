package edgegrid

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

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
	if contract == nil && group == nil {
		return errors.New("function requires at least \"group\" argument")
	}
	if contract == nil && group != nil {
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

	if err = res.BodyJSON(edgeHostnames); err != nil {
		return err
	}

	return nil
}

func (edgeHostnames *PapiEdgeHostnames) FindEdgeHostname(edgeHostname *PapiEdgeHostname) (*PapiEdgeHostname, error) {
	if edgeHostname.DomainSuffix == "" && edgeHostname.EdgeHostnameDomain != "" {
		edgeHostname.DomainSuffix = "edgesuite.net"
		if strings.HasSuffix(edgeHostname.EdgeHostnameDomain, "edgekey.net") {
			edgeHostname.DomainSuffix = "edgekey.net"
		}
	}

	if edgeHostname.DomainPrefix == "" && edgeHostname.EdgeHostnameDomain != "" {
		edgeHostname.DomainPrefix = strings.TrimSuffix(edgeHostname.EdgeHostnameDomain, "."+edgeHostname.DomainSuffix)
	}

	if len(edgeHostnames.EdgeHostnames.Items) == 0 {
		return nil, errors.New("no hostnames found, did you call GetHostnames()?")
	}

	for _, eHn := range edgeHostnames.EdgeHostnames.Items {
		if (eHn.DomainPrefix == edgeHostname.DomainPrefix && eHn.DomainSuffix == edgeHostname.DomainSuffix) || eHn.EdgeHostnameID == edgeHostname.EdgeHostnameID {
			return eHn, nil
		}
	}

	return nil, nil
}

func (edgeHostnames *PapiEdgeHostnames) AddEdgeHostname(edgeHostname *PapiEdgeHostname) {
	found, err := edgeHostnames.FindEdgeHostname(edgeHostname)

	if err != nil || found == nil {
		edgeHostnames.EdgeHostnames.Items = append(edgeHostnames.EdgeHostnames.Items, edgeHostname)
	}

	if err == nil && found != nil && found.EdgeHostnameID == edgeHostname.EdgeHostnameID {
		*found = *edgeHostname
	}
}

// PapiEdgeHostname represents an Edge Hostname resource
type PapiEdgeHostname struct {
	resource
	parent                 *PapiEdgeHostnames
	EdgeHostnameID         string          `json:"edgeHostnameId,omitempty"`
	EdgeHostnameDomain     string          `json:"edgeHostnameDomain,omitempty"`
	ProductID              string          `json:"productId"`
	DomainPrefix           string          `json:"domainPrefix"`
	DomainSuffix           string          `json:"domainSuffix"`
	Status                 PapiStatusValue `json:"status,omitempty"`
	Secure                 bool            `json:"secure,omitempty"`
	IPVersionBehavior      string          `json:"ipVersionBehavior,omitempty"`
	MapDetailsSerialNumber int             `json:"mapDetails:serialNumber,omitempty"`
	MapDetailsSlotNumber   int             `json:"mapDetails:slotNumber,omitempty"`
	MapDetailsMapDomain    string          `json:"mapDetails:mapDomain,omitempty"`
	StatusChange           chan bool       `json:"-"`
}

// NewPapiEdgeHostname creates a new PapiEdgeHostname
func NewPapiEdgeHostname(edgeHostnames *PapiEdgeHostnames) *PapiEdgeHostname {
	edgeHostname := &PapiEdgeHostname{parent: edgeHostnames}
	edgeHostname.Init()
	return edgeHostname
}

func (edgeHostname *PapiEdgeHostname) Init() {
	edgeHostname.Complete = make(chan bool, 1)
	edgeHostname.StatusChange = make(chan bool, 1)
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
		if res.StatusCode == 404 {
			// Check collection for current hostname
			contract := NewPapiContract(NewPapiContracts(edgeHostname.parent.service))
			contract.ContractID = edgeHostname.parent.ContractID
			group := NewPapiGroup(NewPapiGroups(edgeHostname.parent.service))
			group.GroupID = edgeHostname.parent.GroupID

			edgeHostname.parent.GetEdgeHostnames(contract, group, "")
			newEdgeHostname, err := edgeHostname.parent.FindEdgeHostname(edgeHostname)
			if err != nil || newEdgeHostname == nil {
				return NewAPIError(res)
			}

			edgeHostname.EdgeHostnameID = newEdgeHostname.EdgeHostnameID
			edgeHostname.EdgeHostnameDomain = newEdgeHostname.EdgeHostnameDomain
			edgeHostname.ProductID = newEdgeHostname.ProductID
			edgeHostname.DomainPrefix = newEdgeHostname.DomainPrefix
			edgeHostname.DomainSuffix = newEdgeHostname.DomainSuffix
			edgeHostname.Status = newEdgeHostname.Status
			edgeHostname.Secure = newEdgeHostname.Secure
			edgeHostname.IPVersionBehavior = newEdgeHostname.IPVersionBehavior
			edgeHostname.MapDetailsSerialNumber = newEdgeHostname.MapDetailsSerialNumber
			edgeHostname.MapDetailsSlotNumber = newEdgeHostname.MapDetailsSlotNumber
			edgeHostname.MapDetailsMapDomain = newEdgeHostname.MapDetailsMapDomain

			return nil
		}

		return NewAPIError(res)
	}

	newEdgeHostnames := NewPapiEdgeHostnames(edgeHostname.parent.service)
	if err := res.BodyJSON(newEdgeHostnames); err != nil {
		return err
	}

	edgeHostname.EdgeHostnameID = newEdgeHostnames.EdgeHostnames.Items[0].EdgeHostnameID
	edgeHostname.EdgeHostnameDomain = newEdgeHostnames.EdgeHostnames.Items[0].EdgeHostnameDomain
	edgeHostname.ProductID = newEdgeHostnames.EdgeHostnames.Items[0].ProductID
	edgeHostname.DomainPrefix = newEdgeHostnames.EdgeHostnames.Items[0].DomainPrefix
	edgeHostname.DomainSuffix = newEdgeHostnames.EdgeHostnames.Items[0].DomainSuffix
	edgeHostname.Status = newEdgeHostnames.EdgeHostnames.Items[0].Status
	edgeHostname.Secure = newEdgeHostnames.EdgeHostnames.Items[0].Secure
	edgeHostname.IPVersionBehavior = newEdgeHostnames.EdgeHostnames.Items[0].IPVersionBehavior
	edgeHostname.MapDetailsSerialNumber = newEdgeHostnames.EdgeHostnames.Items[0].MapDetailsSerialNumber
	edgeHostname.MapDetailsSlotNumber = newEdgeHostnames.EdgeHostnames.Items[0].MapDetailsSlotNumber
	edgeHostname.MapDetailsMapDomain = newEdgeHostnames.EdgeHostnames.Items[0].MapDetailsMapDomain

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

	// A 404 is returned until the hostname is valid, so just pull the new ID out for now
	url, _ := url.Parse(location["edgeHostnameLink"].(string))
	for _, part := range strings.Split(url.Path, "/") {
		if strings.HasPrefix(part, "ehn_") {
			edgeHostname.EdgeHostnameID = part
		}
	}

	edgeHostname.parent.AddEdgeHostname(edgeHostname)

	return nil
}

// PollStatus will responsibly poll till the property is active or an error occurs
//
// The PapiEdgeHostname.StatusChange is a channel that can be used to
// block on status changes. If a new valid status is returned, true will
// be sent to the channel, otherwise, false will be sent.
//
//	go edgeHostname.PollStatus("")
//	for edgeHostname.Status != edgegrid.PapiStatusActive {
//		select {
//		case statusChanged := <-edgeHostname.StatusChange:
//			if statusChanged == false {
//				break
//			}
//		case <-time.After(time.Minute * 30):
//			break
//		}
//	}
//
//	if edgeHostname.Status == edgegrid.PapiStatusActive {
//		// EdgeHostname activated successfully
//	}
func (edgeHostname *PapiEdgeHostname) PollStatus(options string) bool {
	currentStatus := edgeHostname.Status
	var retry time.Duration = 0
	for currentStatus != PapiStatusActive {
		time.Sleep(retry)
		if retry == 0 {
			retry = time.Minute * 3
		}

		retry -= time.Minute

		err := edgeHostname.GetEdgeHostname(options)
		if err != nil {
			edgeHostname.StatusChange <- false
			return false
		}

		if currentStatus != edgeHostname.Status {
			edgeHostname.StatusChange <- true
		}
		currentStatus = edgeHostname.Status
	}

	return true
}
