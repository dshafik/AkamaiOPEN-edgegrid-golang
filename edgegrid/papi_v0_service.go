package edgegrid

type PapiV0Service struct {
	client *Client
	config *Config
}

func NewPapiV0Service(client *Client, config *Config) *PapiV0Service {
	service := &PapiV0Service{client: client, config: config}
	return service
}

func (papi *PapiV0Service) GetGroups() (*PapiGroups, error) {
	res, err := papi.client.Get("/papi/v0/groups")
	if err != nil {
		return nil, err
	}

	if res.IsError() == true {
		return nil, NewApiError(res)
	}

	groups := &PapiGroups{service: papi}
	err = res.BodyJson(&groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (papi *PapiV0Service) GetContracts() (*PapiContracts, error) {
	res, err := papi.client.Get("/papi/v0/contracts")
	if err != nil {
		return nil, err
	}

	if res.IsError() == true {
		return nil, NewApiError(res)
	}

	contracts := &PapiContracts{service: papi}
	err = res.BodyJson(&contracts)
	if err != nil {
		return nil, err
	}

	return contracts, nil
}
