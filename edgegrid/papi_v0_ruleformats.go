package edgegrid

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

	newRuleFormats := NewPapiRuleFormats(ruleFormats.service)
	if err := res.BodyJSON(newRuleFormats); err != nil {
		return err
	}

	*ruleFormats = *newRuleFormats

	return nil
}
