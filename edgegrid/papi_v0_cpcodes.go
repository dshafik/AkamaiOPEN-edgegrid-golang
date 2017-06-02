package edgegrid

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

// PapiCpCodes represents a collection of CP Codes
//
// See: PapiCpCodes.GetCpCodes()
// API Docs: https://developer.akamai.com/api/luna/papi/data.html#cpcode
type PapiCpCodes struct {
	resource
	service    *PapiV0Service
	AccountID  string        `json:"accountId"`
	Contract   *PapiContract `json:"-"`
	ContractID string        `json:"contractId"`
	GroupID    string        `json:"groupId"`
	Group      *PapiGroup    `json:"-"`
	CpCodes    struct {
		Items []*PapiCpCode `json:"items"`
	} `json:"cpcodes"`
}

// NewPapiCpCodes creates a new *PapiCpCodes
func NewPapiCpCodes(service *PapiV0Service, contract *PapiContract, group *PapiGroup) *PapiCpCodes {
	return &PapiCpCodes{
		service:  service,
		Contract: contract,
		Group:    group,
	}
}

// PostUnmarshalJSON is called after UnmarshalJSON to setup the
// structs internal state. The cpcodes.Complete channel is utilized
// to communicate full completion.
func (cpcodes *PapiCpCodes) PostUnmarshalJSON() error {
	cpcodes.Init()

	cpcodes.Contract = NewPapiContract(NewPapiContracts(cpcodes.service))
	cpcodes.Contract.ContractID = cpcodes.ContractID

	cpcodes.Group = NewPapiGroup(NewPapiGroups(cpcodes.service))
	cpcodes.Group.GroupID = cpcodes.GroupID

	go cpcodes.Group.GetGroup()
	go cpcodes.Contract.GetContract()

	go (func(cpcodes *PapiCpCodes) {
		contractComplete := <-cpcodes.Contract.Complete
		groupComplete := <-cpcodes.Group.Complete
		cpcodes.Complete <- (contractComplete && groupComplete)
	})(cpcodes)

	for key, cpcode := range cpcodes.CpCodes.Items {
		cpcodes.CpCodes.Items[key].parent = cpcodes

		if cpcode, ok := json.ImplementsPostJSONUnmarshaler(cpcode); ok {
			if err := cpcode.(json.PostJSONUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetCpCodes populates a *PapiCpCodes with it's related CP Codes
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listcpcodes
// Endpoint: GET /papi/v0/cpcodes/{?contractId,groupId}
func (cpcodes *PapiCpCodes) GetCpCodes() error {
	if cpcodes.Contract == nil {
		cpcodes.Contract = NewPapiContract(NewPapiContracts(cpcodes.service))
		cpcodes.Contract.ContractID = cpcodes.Group.ContractIDs[0]
	}
	res, err := cpcodes.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/cpcodes?groupId=%s&contractId=%s",
			cpcodes.Group.GroupID,
			cpcodes.Contract.ContractID,
		),
	)
	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newCpcodes := NewPapiCpCodes(cpcodes.service, nil, nil)
	if err = res.BodyJSON(newCpcodes); err != nil {
		return err
	}

	*cpcodes = *newCpcodes

	return nil
}

func (cpcodes *PapiCpCodes) FindCpCode(nameOrId string) (*PapiCpCode, error) {
	if len(cpcodes.CpCodes.Items) == 0 {
		err := cpcodes.GetCpCodes()
		if err != nil {
			return nil, err
		}

		if len(cpcodes.CpCodes.Items) == 0 {
			return nil, fmt.Errorf("unable to fetch CP codes for group/contract")
		}
	}

	for _, cpcode := range cpcodes.CpCodes.Items {
		if cpcode.CpcodeID == nameOrId || cpcode.CpcodeID == "cpc_"+nameOrId || cpcode.CpcodeName == nameOrId {
			return cpcode, nil
		}
	}

	return nil, nil
}

// NewCpCode creates a new *PapiCpCode associated with this *PapiCpCodes as it's parent.
func (cpcodes *PapiCpCodes) NewCpCode() *PapiCpCode {
	cpcode := NewPapiCpCode(cpcodes)
	cpcodes.CpCodes.Items = append(cpcodes.CpCodes.Items, cpcode)
	return cpcode
}

// PapiCpCode represents a single CP Code
//
// API Docs: https://developer.akamai.com/api/luna/papi/data.html#cpcode
type PapiCpCode struct {
	resource
	parent      *PapiCpCodes
	CpcodeID    string    `json:"cpcodeId,omitempty"`
	CpcodeName  string    `json:"cpcodeName"`
	ProductID   string    `json:"productId,omitempty"`
	ProductIDs  []string  `json:"productIds,omitempty"`
	CreatedDate time.Time `json:"createdDate,omitempty"`
}

// NewPapiCpCode creates a new *PapiCpCode
func NewPapiCpCode(parent *PapiCpCodes) *PapiCpCode {
	cpcode := &PapiCpCode{parent: parent}
	cpcode.Init()
	return cpcode
}

// GetCpCode populates the *PapiCpCode with it's data
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getacpcode
// Endpoint: GET /papi/v0/cpcodes/{cpcodeId}{?contractId,groupId}
func (cpcode *PapiCpCode) GetCpCode() error {
	res, err := cpcode.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/cpcodes/%s?contractId=%s&groupId=%s",
			cpcode.CpcodeID,
			cpcode.parent.Contract.ContractID,
			cpcode.parent.Group.GroupID,
		),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newCpcodes := NewPapiCpCodes(cpcode.parent.service, nil, nil)
	if err = res.BodyJSON(newCpcodes); err != nil {
		return err
	}
	if len(newCpcodes.CpCodes.Items) == 0 {
		return fmt.Errorf("CP Code \"%s\" not found", cpcode.CpcodeID)
	}

	*cpcode = *newCpcodes.CpCodes.Items[0]
	return nil
}

// ID retrieves a CP Codes integer ID
//
// PAPI Behaviors require the integer ID, rather than the prefixed string returned
func (cpcode *PapiCpCode) ID() int {
	id, err := strconv.Atoi(strings.TrimPrefix(cpcode.CpcodeID, "cpc_"))
	if err != nil {
		return 0
	}

	return id
}

// Save will create a new CP Code. You cannot update a CP Code;
// trying to do so will result in an error.
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#createanewcpcode
// Endpoint: POST /papi/v0/cpcodes/{?contractId,groupId}
func (cpcode *PapiCpCode) Save() error {
	res, err := cpcode.parent.service.client.PostJSON(
		fmt.Sprintf(
			"/papi/v0/cpcodes?contractId=%s&groupId=%s",
			cpcode.parent.Contract.ContractID,
			cpcode.parent.Group.GroupID,
		),
		JSONBody{"productId": cpcode.ProductID, "productIds": cpcode.ProductIDs, "cpcodeName": cpcode.CpcodeName},
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

	res, err = cpcode.parent.service.client.Get(
		location["cpcodeLink"].(string),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	cpcodes := NewPapiCpCodes(cpcode.parent.service, nil, nil)
	if err != nil {
		return err
	}

	if err = res.BodyJSON(cpcodes); err != nil {
		return err
	}

	newCpcode := cpcodes.CpCodes.Items[0]
	newCpcode.parent = cpcode.parent
	cpcode.parent.CpCodes.Items = append(cpcode.parent.CpCodes.Items, newCpcode)

	*cpcode = *newCpcode

	return nil
}
