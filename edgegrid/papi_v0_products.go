package edgegrid

import "github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"

type PapiProducts struct {
	service    *PapiV0Service
	AccountId  string `json:"accountId"`
	ContractId string `json:"contractId"`
	Products   struct {
		Items []*PapiProduct `json:"items"`
	} `json:"products"`
	Complete chan bool `json:"-"`
}

func NewPapiProducts(service *PapiV0Service) *PapiProducts {
	products := &PapiProducts{service: service}
	products.Init()

	return products
}

func (products *PapiProducts) Init() {
	products.Complete = make(chan bool, 1)
}

func (products *PapiProducts) PostUnmarshalJSON() error {
	for key, product := range products.Products.Items {
		products.Products.Items[key].parent = products
		if product, ok := json.ImplementsPostJsonUnmarshaler(product); ok {
			product.(json.PostJsonUnmarshaler).PostUnmarshalJSON()
		}
	}

	return nil
}

type PapiProduct struct {
	parent      *PapiProducts
	ProductName string    `json:"productName"`
	ProductId   string    `json:"productId"`
	Complete    chan bool `json:"-"`
}

func NewPapiProduct(parent *PapiProducts) *PapiProduct {
	product := &PapiProduct{parent: parent}
	product.Init()

	return product
}

func (product *PapiProduct) Init() {
	product.Complete = make(chan bool, 1)
}

func (product *PapiProduct) PostUnmarshalJSON() error {
	product.Init()
	product.Complete <- true
	return nil
}
