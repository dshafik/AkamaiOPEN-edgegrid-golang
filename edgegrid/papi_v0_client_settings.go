package edgegrid

type PapiClientSettings struct {
	Resource
	service    *PapiV0Service
	RuleFormat string `json:"ruleFormat"`
}

func NewPapiClientSettings(service *PapiV0Service) *PapiClientSettings {
	clientSettings := &PapiClientSettings{service: service}
	clientSettings.Init()

	return clientSettings
}
