package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

// PapiProducts represents a collection of products
type PapiProducts struct {
	resource
	service    *PapiV0Service
	AccountID  string `json:"accountId"`
	ContractID string `json:"contractId"`
	Products   struct {
		Items []*PapiProduct `json:"items"`
	} `json:"products"`
}

// NewPapiProducts creates a new PapiProducts
func NewPapiProducts(service *PapiV0Service) *PapiProducts {
	products := &PapiProducts{service: service}
	products.Init()

	return products
}

// PostUnmarshalJSON is called after JSON unmarshaling into PapiEdgeHostnames
//
// See: edgegrid/json.Unmarshal()
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

	if err = res.BodyJSON(products); err != nil {
		return err
	}

	return nil
}

// PapiProduct represents a product resource
type PapiProduct struct {
	resource
	parent      *PapiProducts
	ProductName string `json:"productName"`
	ProductID   string `json:"productId"`
}

// NewPapiProduct creates a new PapiProduct
func NewPapiProduct(parent *PapiProducts) *PapiProduct {
	product := &PapiProduct{parent: parent}
	product.Init()

	return product
}
