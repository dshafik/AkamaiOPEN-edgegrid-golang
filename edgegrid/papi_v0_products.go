package edgegrid

import "github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"

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
			product.(json.PostJsonUnmarshaler).PostUnmarshalJSON()
		}
	}

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
