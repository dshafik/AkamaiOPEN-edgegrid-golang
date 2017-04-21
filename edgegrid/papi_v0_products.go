package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

type PapiProducts struct {
	Resource
	service    *PapiV0Service
	AccountId  string `json:"accountId"`
	ContractId string `json:"contractId"`
	Products   struct {
		Items []*PapiProduct `json:"items"`
	} `json:"products"`
}

func NewPapiProducts(service *PapiV0Service) *PapiProducts {
	products := &PapiProducts{service: service}
	products.Init()

	return products
}

func (products *PapiProducts) PostUnmarshalJSON() error {
	products.Init()

	for key, product := range products.Products.Items {
		products.Products.Items[key].parent = products
		if product, ok := json.ImplementsPostJsonUnmarshaler(product); ok {
			if err := product.(json.PostJsonUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (products *PapiProducts) GetProducts(contract *PapiContract) error {
	res, err := products.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/products?contractId=%s",
			contract.ContractId,
		),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	newProducts := NewPapiProducts(products.service)
	if err = res.BodyJson(products); err != nil {
		return err
	}

	*products = *newProducts

	return nil
}

type PapiProduct struct {
	Resource
	parent      *PapiProducts
	ProductName string `json:"productName"`
	ProductId   string `json:"productId"`
}

func NewPapiProduct(parent *PapiProducts) *PapiProduct {
	product := &PapiProduct{parent: parent}
	product.Init()

	return product
}
