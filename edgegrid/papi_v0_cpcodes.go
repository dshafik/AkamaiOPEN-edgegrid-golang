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
	AccountId  string        `json:"accountId"`
	Contract   *PapiContract `json:"-"`
	ContractId string        `json:"contractId"`
	GroupId    string        `json:"groupId"`
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

	cpcodes.Contract = NewPapiContract(&PapiContracts{service: cpcodes.service})
	cpcodes.Contract.ContractId = cpcodes.ContractId

	cpcodes.Group = NewPapiGroup(&PapiGroups{service: cpcodes.service})
	cpcodes.Group.GroupId = cpcodes.GroupId

	go cpcodes.Group.GetGroup()
	go cpcodes.Contract.GetContract()

	go (func(cpcodes *PapiCpCodes) {
		contractComplete := <-cpcodes.Contract.Complete
		groupComplete := <-cpcodes.Group.Complete
		cpcodes.Complete <- (contractComplete && groupComplete)
	})(cpcodes)

	for key, cpcode := range cpcodes.CpCodes.Items {
		cpcodes.CpCodes.Items[key].parent = cpcodes

		if cpcode, ok := json.ImplementsPostJsonUnmarshaler(cpcode); ok {
			if err := cpcode.(json.PostJsonUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cpcodes *PapiCpCodes) GetCpCodes(contract *PapiContract, group *PapiGroup) error {
	if contract == nil {
		contract = NewPapiContract(NewPapiContracts(cpcodes.service))
		contract.ContractId = group.ContractIds[0]
	}
	res, err := cpcodes.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/cpcodes?groupId=%s&contractId=%s",
			group.GroupId,
			contract.ContractId,
		),
	)
	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	newCpcodes := NewPapiCpCodes(cpcodes.service)
	err = res.BodyJson(newCpcodes)
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
	CpcodeId    string    `json:"cpcodeId,omitempty"`
	CpcodeName  string    `json:"cpcodeName"`
	ProductId   string    `json:"productId,omitempty"`
	ProductIds  []string  `json:"productIds,omitempty"`
	CreatedDate time.Time `json:"createdDate,omitempty"`
}

func NewPapiCpCode(parent *PapiCpCodes) *PapiCpCode {
	cpcode := &PapiCpCode{parent: parent}
	cpcode.Init()
	return cpcode
}

func (cpcode *PapiCpCode) Id() int {
	id, err := strconv.Atoi(strings.TrimPrefix(cpcode.CpcodeId, "cpc_"))
	if err != nil {
		return 0
	}

	return id
}

func (cpcode *PapiCpCode) Save() error {
	res, err := cpcode.parent.service.client.PostJson(
		fmt.Sprintf(
			"/papi/v0/cpcodes?contractId=%s&groupId=%s",
			cpcode.parent.Contract.ContractId,
			cpcode.parent.Group.GroupId,
		),
		map[string]interface{}{"productId": cpcode.ProductId, "productIds": cpcode.ProductIds, "cpcodeName": cpcode.CpcodeName},
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	var location map[string]interface{}
	res.BodyJson(&location)

	res, err = cpcode.parent.service.client.Get(
		location["cpcodeLink"].(string),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	cpcodes := NewPapiCpCodes(cpcode.parent.service)
	err = res.BodyJson(cpcodes)
	if err != nil {
		return err
	}

	newCpcode := cpcodes.CpCodes.Items[0]
	newCpcode.parent = cpcode.parent
	cpcode.parent.CpCodes.Items = append(cpcode.parent.CpCodes.Items, newCpcode)

	*cpcode = *newCpcode

	return nil
}
