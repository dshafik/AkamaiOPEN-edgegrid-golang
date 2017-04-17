package edgegrid

type PapiClientSettings struct {
	service    *PapiV0Service
	RuleFormat string    `json:"ruleFormat"`
	Complete   chan bool `json:"-"`
}

func NewPapiClientSettings(service *PapiV0Service) *PapiClientSettings {
	clientSettings := &PapiClientSettings{service: service}
	clientSettings.Init()

	return clientSettings
}

func (clientSettings *PapiClientSettings) Init() {
	clientSettings.Complete = make(chan bool, 1)
}

func (clientSettings *PapiClientSettings) PostUnmashalJSON() error {
	clientSettings.Init()
	clientSettings.Complete <- true

	return nil
}
