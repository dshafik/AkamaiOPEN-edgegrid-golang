package edgegrid

// PapiClientSettings represents the PAPI client settings resource
type PapiClientSettings struct {
	resource
	service    *PapiV0Service
	RuleFormat string `json:"ruleFormat"`
}

// NewPapiClientSettings creates a new PapiClientSettings
func NewPapiClientSettings(service *PapiV0Service) *PapiClientSettings {
	clientSettings := &PapiClientSettings{service: service}
	clientSettings.Init()

	return clientSettings
}

// GetClientSettings populates PapiClientSettings
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getclientsettings
// Endpoint: GET /papi/v0/client-settings
func (clientSettings *PapiClientSettings) GetClientSettings() error {
	res, err := clientSettings.service.client.Get("/papi/v0/client-settings")

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newClientSettings := NewPapiClientSettings(clientSettings.service)
	if err := res.BodyJSON(newClientSettings); err != nil {
		return err
	}

	*clientSettings = *newClientSettings

	return nil
}

// Save updates client settings
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#updateclientsettings
// Endpoint: PUT /papi/v0/client-settings
func (clientSettings *PapiClientSettings) Save() error {
	res, err := clientSettings.service.client.PutJSON(
		"/papi/v0/client-settings",
		clientSettings,
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newClientSettings := NewPapiClientSettings(clientSettings.service)
	if err := res.BodyJSON(newClientSettings); err != nil {
		return err
	}

	*clientSettings = *newClientSettings

	return nil
}
