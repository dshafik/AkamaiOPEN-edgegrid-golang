package edgegrid

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
)

// PapiRuleFormats is a collection of available rule formats
type PapiRuleFormats struct {
	resource
	service     *PapiV0Service
	RuleFormats struct {
		Items []string `json:"items"`
	} `json:"ruleFormats"`
}

// NewPapiRuleFormats creates a new PapiRuleFormats
func NewPapiRuleFormats(service *PapiV0Service) *PapiRuleFormats {
	ruleFormats := &PapiRuleFormats{service: service}
	ruleFormats.Init()

	return ruleFormats
}

// GetRuleFormats populates PapiRuleFormats
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listruleformats
// Endpoint: GET /papi/v0/rule-formats
func (ruleFormats *PapiRuleFormats) GetRuleFormats() error {
	res, err := ruleFormats.service.client.Get("/papi/v0/rule-formats")

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	if err := res.BodyJSON(ruleFormats); err != nil {
		return err
	}

	return nil
}

// GetSchema fetches the schema for a given product and rule format
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getaruleformatsschema
// Endpoint: /papi/v0/schemas/products/{productId}/{ruleFormat}
func (ruleFormats *PapiRuleFormats) GetSchema(product string, ruleFormat string) (*gojsonschema.Schema, error) {
	res, err := ruleFormats.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/schemas/products/%s/%s",
			product,
			ruleFormat,
		),
	)

	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, NewAPIError(res)
	}

	schemaBytes, _ := ioutil.ReadAll(res.Body)
	schemaBody := string(schemaBytes)
	loader := gojsonschema.NewStringLoader(schemaBody)
	schema, err := gojsonschema.NewSchema(loader)

	return schema, err
}
