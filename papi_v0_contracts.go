package edgegrid

import (
	"encoding/json"
	"fmt"
)

type PapiContracts struct {
	service   *PapiV0Service
	AccountId string `json:"accountId"`
	Contracts struct {
		Items []*PapiContract `json:"items"`
	} `json:"contracts"`
}

func (contracts *PapiContracts) UnmarshalJSON(b []byte) error {
	type PapiContractsTemp PapiContracts
	temp := &PapiContractsTemp{service: contracts.service}

	if err := json.Unmarshal(b, temp); err != nil {
		return err
	}
	*contracts = (PapiContracts)(*temp)

	for key, _ := range contracts.Contracts.Items {
		contracts.Contracts.Items[key].parent = contracts
	}

	return nil
}

type PapiContract struct {
	parent           *PapiContracts
	ContractId       string `json:"contractId"`
	ContractTypeName string `json:"contractTypeName"`
}

func (contract *PapiContract) GetProducts() (*PapiProducts, error) {
	res, err := contract.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/products?contractId=",
			contract.ContractId,
		),
	);

	if err != nil {
		return nil, err
	}

	if res.IsError() == true {
		return nil, NewApiError(res)
	}

	products := &PapiProducts{service: contract.parent.service}
	err = res.BodyJson(&products)
	if err != nil {
		return nil, err
	}

	return products, nil
}