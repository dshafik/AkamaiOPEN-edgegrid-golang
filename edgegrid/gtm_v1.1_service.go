package edgegrid

type GtmV11Service struct {
	client *Client
	config *Config
}

func NewGtmV11Service(client *Client, config *Config) *GtmV11Service {
	service := &GtmV11Service{client: client, config: config}
	return service
}

func (gtm *GtmV11Service) GetDomains() (*GtmDomains, error) {
	res, error := gtm.client.Get("/config-gtm/v1/domains")
	if error != nil {
		return nil, error
	}

	if res.IsError() {
		return nil, NewApiError(res)
	}

	domains := &GtmDomains{service: gtm}
	err := res.BodyJson(&domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (gtm *GtmV11Service) GetDomain(domainName string) (*GtmDomain, error) {
	domains := &GtmDomains{service: gtm}
	return domains.GetDomain(domainName)
}

type GtmHypermediaLinks struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}
