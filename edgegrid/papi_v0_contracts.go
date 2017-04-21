package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

type PapiContracts struct {
	Resource
	service   *PapiV0Service
	AccountId string `json:"accountId"`
	Contracts struct {
		Items []*PapiContract `json:"items"`
	} `json:"contracts"`
}

func NewPapiContracts(service *PapiV0Service) *PapiContracts {
	contracts := &PapiContracts{
		service: service,
	}
	contracts.Init()
	return contracts
}

func (contracts *PapiContracts) PostUnmarshalJSON() error {
	contracts.Init()

	for key, contract := range contracts.Contracts.Items {
		contracts.Contracts.Items[key].parent = contracts

		if contract, ok := json.ImplementsPostJsonUnmarshaler(contract); ok {
			if err := contract.(json.PostJsonUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}
	contracts.Complete <- true

	return nil
}

func (contracts *PapiContracts) GetContracts() error {
	res, err := contracts.service.client.Get("/papi/v0/contracts")
	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	newContracts := NewPapiContracts(contracts.service)
	if err = res.BodyJson(newContracts); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	*contracts = *newContracts

	return nil
}

type PapiContract struct {
	Resource
	parent           *PapiContracts
	ContractId       string `json:"contractId"`
	ContractTypeName string `json:"contractTypeName"`
}

func NewPapiContract(parent *PapiContracts) *PapiContract {
	contract := &PapiContract{
		parent: parent,
	}
	contract.Init()
	return contract
}

func (contract *PapiContract) GetContract() {
	contracts, err := contract.parent.service.GetContracts()
	if err != nil {
		return
	}

	for _, c := range contracts.Contracts.Items {
		if c.ContractId == contract.ContractId {
			contract.parent = c.parent
			contract.ContractTypeName = c.ContractTypeName
			contract.Complete <- true
			return
		}
	}
	contract.Complete <- false
}

func (contract *PapiContract) GetProducts() (*PapiProducts, error) {
	res, err := contract.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/products?contractId=",
			contract.ContractId,
		),
	)

	if err != nil {
		return nil, err
	}

	if res.IsError() == true {
		return nil, NewApiError(res)
	}

	products := NewPapiProducts(contract.parent.service)
	err = res.BodyJson(products)
	if err != nil {
		return nil, err
	}

	return products, nil
}
