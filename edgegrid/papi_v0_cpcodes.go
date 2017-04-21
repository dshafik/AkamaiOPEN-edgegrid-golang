package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
	"strconv"
	"strings"
	"time"
)

type PapiCpCodes struct {
	Resource
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

func NewPapiCpCodes(service *PapiV0Service) *PapiCpCodes {
	cpcodes := &PapiCpCodes{service: service}
	return cpcodes
}

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

func (cpcodes *PapiCpCodes) GetCpCodes(contract *PapiContract, group *PapiGroup) error {
	if contract == nil {
		contract = NewPapiContract(NewPapiContracts(cpcodes.service))
		contract.ContractID = group.ContractIDs[0]
	}
	res, err := cpcodes.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/cpcodes?groupId=%s&contractId=%s",
			group.GroupID,
			contract.ContractID,
		),
	)
	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newCpcodes := NewPapiCpCodes(cpcodes.service)
	err = res.BodyJSON(newCpcodes)
	if err != nil {
		return err
	}

	*cpcodes = *newCpcodes

	return nil
}

func (cpcodes *PapiCpCodes) NewCpCode() *PapiCpCode {
	return NewPapiCpCode(cpcodes)
}

type PapiCpCode struct {
	Resource
	parent      *PapiCpCodes
	CpcodeID    string    `json:"cpcodeId,omitempty"`
	CpcodeName  string    `json:"cpcodeName"`
	ProductID   string    `json:"productId,omitempty"`
	ProductIDs  []string  `json:"productIds,omitempty"`
	CreatedDate time.Time `json:"createdDate,omitempty"`
}

func NewPapiCpCode(parent *PapiCpCodes) *PapiCpCode {
	cpcode := &PapiCpCode{parent: parent}
	cpcode.Init()
	return cpcode
}

func (cpcode *PapiCpCode) ID() int {
	id, err := strconv.Atoi(strings.TrimPrefix(cpcode.CpcodeID, "cpc_"))
	if err != nil {
		return 0
	}

	return id
}

func (cpcode *PapiCpCode) Save() error {
	res, err := cpcode.parent.service.client.PostJSON(
		fmt.Sprintf(
			"/papi/v0/cpcodes?contractId=%s&groupId=%s",
			cpcode.parent.Contract.ContractID,
			cpcode.parent.Group.GroupID,
		),
		map[string]interface{}{"productId": cpcode.ProductID, "productIds": cpcode.ProductIDs, "cpcodeName": cpcode.CpcodeName},
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	var location map[string]interface{}
	if err := res.BodyJSON(&location); err != nil {
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

	cpcodes := NewPapiCpCodes(cpcode.parent.service)
	err = res.BodyJSON(cpcodes)
	if err != nil {
		return err
	}

	newCpcode := cpcodes.CpCodes.Items[0]
	newCpcode.parent = cpcode.parent
	cpcode.parent.CpCodes.Items = append(cpcode.parent.CpCodes.Items, newCpcode)

	*cpcode = *newCpcode

	return nil
}
