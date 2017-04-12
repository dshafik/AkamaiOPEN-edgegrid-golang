package edgegrid

import "encoding/json"

type PapiProducts struct {
	service    *PapiV0Service
	AccountId  string `json:"accountId"`
	ContractId string `json:"contractId"`
	Products   struct {
		Items []*PapiProduct `json:"items"`
	} `json:"products"`
}

func (products *PapiProducts) UnmarshalJSON(b []byte) error {
	type PapiProductsTemp PapiProducts
	temp := &PapiProductsTemp{service: products.service}

	if err := json.Unmarshal(b, temp); err != nil {
		return err
	}
	*products = (PapiProducts)(*temp)

	for key, _ := range products.Products.Items {
		products.Products.Items[key].parent = products
	}

	return nil
}

type PapiProduct struct {
	parent      *PapiProducts
	ProductName string `json:"productName"`
	ProductId   string `json:"productId"`
}
