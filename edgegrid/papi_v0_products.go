package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

type PapiProducts struct {
	resource
	service    *PapiV0Service
	AccountID  string `json:"accountId"`
	ContractID string `json:"contractId"`
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
		if product, ok := json.ImplementsPostJSONUnmarshaler(product); ok {
			if err := product.(json.PostJSONUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetProducts populates PapiProducts with product data
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listproducts
// Endpoint: GET /papi/v0/products/{?contractId}
func (products *PapiProducts) GetProducts(contract *PapiContract) error {
	res, err := products.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/products?contractId=%s",
			contract.ContractID,
		),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newProducts := NewPapiProducts(products.service)
	if err = res.BodyJSON(newProducts); err != nil {
		return err
	}

	*products = *newProducts

	return nil
}

type PapiProduct struct {
	resource
	parent      *PapiProducts
	ProductName string `json:"productName"`
	ProductID   string `json:"productId"`
}

func NewPapiProduct(parent *PapiProducts) *PapiProduct {
	product := &PapiProduct{parent: parent}
	product.Init()

	return product
}
