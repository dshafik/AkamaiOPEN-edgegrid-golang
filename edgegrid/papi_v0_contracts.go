package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

type PapiContracts struct {
	service   *PapiV0Service
	AccountId string `json:"accountId"`
	Contracts struct {
		Items []*PapiContract `json:"items"`
	} `json:"contracts"`
	Complete chan bool `json:"-"`
}

func NewPapiContracts(service *PapiV0Service) *PapiContracts {
	contracts := &PapiContracts{
		service: service,
	}
	contracts.Init()
	return contracts
}

func (contracts *PapiContracts) Init() {
	contracts.Complete = make(chan bool, 1)
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

type PapiContract struct {
	parent           *PapiContracts
	ContractId       string    `json:"contractId"`
	ContractTypeName string    `json:"contractTypeName"`
	Complete         chan bool `json:"-"`
}

func NewPapiContract(parent *PapiContracts) *PapiContract {
	contract := &PapiContract{
		parent: parent,
	}
	contract.Init()
	return contract
}

func (contract *PapiContract) Init() {
	contract.Complete = make(chan bool, 1)
}

func (contract *PapiContract) PostUnmarshalJSON() error {
	contract.Init()
	contract.Complete <- true
	return nil
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

	products := &PapiProducts{service: contract.parent.service}
	err = res.BodyJson(products)
	if err != nil {
		return nil, err
	}

	return products, nil
}
