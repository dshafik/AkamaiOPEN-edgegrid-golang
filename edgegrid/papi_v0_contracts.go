package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

// PapiContracts represents a collection of property manager contracts
type PapiContracts struct {
	resource
	service   *PapiV0Service
	AccountID string `json:"accountId"`
	Contracts struct {
		Items []*PapiContract `json:"items"`
	} `json:"contracts"`
}

// NewPapiContracts creates a new PapiContracts
func NewPapiContracts(service *PapiV0Service) *PapiContracts {
	contracts := &PapiContracts{
		service: service,
	}
	contracts.Init()
	return contracts
}

// PostUnmarshalJSON is called after JSON unmarshaling into PapiEdgeHostnames
//
// See: edgegrid/json.Unmarshal()
func (contracts *PapiContracts) PostUnmarshalJSON() error {
	contracts.Init()

	for key, contract := range contracts.Contracts.Items {
		contracts.Contracts.Items[key].parent = contracts

		if contract, ok := json.ImplementsPostJSONUnmarshaler(contract); ok {
			if err := contract.(json.PostJSONUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}
	contracts.Complete <- true

	return nil
}

// GetContracts populates PapiContracts with contract data
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listcontracts
// Endpoint: GET /papi/v0/contracts
func (contracts *PapiContracts) GetContracts() error {
	res, err := contracts.service.client.Get("/papi/v0/contracts")
	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newContracts := NewPapiContracts(contracts.service)
	if err = res.BodyJSON(newContracts); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	*contracts = *newContracts

	return nil
}

// PapiContract represents a property contract resource
type PapiContract struct {
	resource
	parent           *PapiContracts
	ContractID       string `json:"contractId"`
	ContractTypeName string `json:"contractTypeName"`
}

// NewPapiContract creates a new PapiContract
func NewPapiContract(parent *PapiContracts) *PapiContract {
	contract := &PapiContract{
		parent: parent,
	}
	contract.Init()
	return contract
}

// GetContract populates a PapiContract
func (contract *PapiContract) GetContract() error {
	contracts, err := contract.parent.service.GetContracts()
	if err != nil {
		return err
	}

	for _, c := range contracts.Contracts.Items {
		if c.ContractID == contract.ContractID {
			contract.parent = c.parent
			contract.ContractTypeName = c.ContractTypeName
			contract.Complete <- true
			return nil
		}
	}
	contract.Complete <- false
	return fmt.Errorf("contract \"%s\" not found", contract.ContractID)
}

// GetProducts gets products associated with a contract
func (contract *PapiContract) GetProducts() (*PapiProducts, error) {
	res, err := contract.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/products?contractId=%s",
			contract.ContractID,
		),
	)

	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, NewAPIError(res)
	}

	products := NewPapiProducts(contract.parent.service)
	if err = res.BodyJSON(products); err != nil {
		return nil, err
	}

	return products, nil
}
