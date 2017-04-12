package edgegrid

import (
	"fmt"
)

type GtmDomains struct {
	service *GtmV11Service
	Items   []*GtmDomain `json:"items"`
}

func (domains *GtmDomains) GetDomain(domainName string) (*GtmDomain, error) {
	if len(domains.Items) > 0 {
		for _, domain := range domains.Items {
			if domain.domain == domainName {
				return domain, nil
			}
		}
	}

	res, err := domains.service.client.Get(fmt.Sprintf("/config-gtm/v1/domains/%s", domainName))

	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, NewApiError(res)
	}

	domain := &GtmDomain{parent: domains, domain: domainName}
	err = res.BodyJson(&domain)

	if err != nil {
		return nil, err
	}

	return domain, nil
}

func (domains *GtmDomains) NewDomain(domainName string) *GtmDomain {
	return &GtmDomain{parent: domains, domain: domainName}
}

type GtmDomain struct {
	parent                      *GtmDomains
	domain                      string
	CidrMaps                    []*GtmCidrMap       `json:"cidrMaps"`
	Datacenters                 []*GtmDatacenter    `json:"datacenters"`
	DefaultSslClientCertificate string              `json:"defaultSslClientCertificate"`
	DefaultSslClientPrivateKey  string              `json:"defaultSslClientPrivateKey"`
	DefaultUnreachableThreshold string              `json:"defaultUnreachableThreshold"`
	EmailNotificationList       []string            `json:"emailNotificationList"`
	GeographicMaps              []*GtmGeographicMap `json:"geographicMaps"`
	LastModified                string              `json:"lastModified"`
	LastModifiedBy              string              `json:"lastModifiedBy"`
	LoadFeedback                bool                `json:"loadFeedback"`
	LoadImbalancePercentage     float64             `json:"loadImbalancePercentage"`
	MinPingableRegionFraction   interface{}         `json:"minPingableRegionFraction"`
	ModificationComments        string              `json:"modificationComments"`
	Name                        string              `json:"name"`
	PingInterval                interface{}         `json:"pingInterval"`
	Properties                  []*GtmProperty      `json:"properties"`
	Resources                   []*GtmResource      `json:"resources"`
	RoundRobinPrefix            interface{}         `json:"roundRobinPrefix"`
	ServermonitorLivenessCount  interface{}         `json:"servermonitorLivenessCount"`
	ServermonitorLoadCount      interface{}         `json:"servermonitorLoadCount"`
	Status                      *GtmDomainStatus    `json:"status"`
	Type                        string              `json:"type"`

	Links []*GtmHypermediaLinks `json:"links,omitempty"`
}

type GtmDomainStatus struct {
	ChangeID              string                `json:"changeId"`
	Message               string                `json:"message"`
	PassingValidation     bool                  `json:"passingValidation"`
	PropagationStatus     string                `json:"propagationStatus"`
	PropagationStatusDate string                `json:"propagationStatusDate"`
	Links                 []*GtmHypermediaLinks `json:"links"`
}

func (domain *GtmDomain) Save() error {
	req, err := domain.parent.service.client.NewJsonRequest("PUT", fmt.Sprintf("/config-gtm/v1/domains/%s", domain.domain), domain)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/vnd.config-gtm.v1.1+json")

	res, err := domain.parent.service.client.Do(req)

	if res.IsError() {
		return NewApiError(res)
	}

	return nil
}

func (domain *GtmDomain) GetDatacenters() (*GtmDatacenters, error) {
	datacenters := &GtmDatacenters{service: domain.parent.service}
	if err := domain.getDomainItems("datacenters", &datacenters); err != nil {
		return nil, err
	}

	return datacenters, nil
}

func (domain *GtmDomain) GetProperties() (*GtmProperties, error) {
	properties := &GtmProperties{service: domain.parent.service}
	if err := domain.getDomainItems("properties", &properties); err != nil {
		return nil, err
	}

	return properties, nil
}

func (domain *GtmDomain) GetGeographicMaps() (*GtmGeographicMaps, error) {
	geographicMaps := &GtmGeographicMaps{service: domain.parent.service}
	if err := domain.getDomainItems("geographic-maps", &geographicMaps); err != nil {
		return nil, err
	}

	return geographicMaps, nil
}

func (domain *GtmDomain) GetCidrMaps() (*GtmCidrMaps, error) {
	cidrMaps := &GtmCidrMaps{service: domain.parent.service}
	if err := domain.getDomainItems("cidr-maps", &cidrMaps); err != nil {
		return nil, err
	}

	return cidrMaps, nil
}

func (domain *GtmDomain) GetAutonomousSystemMaps() (*GtmAutonomousSystemMaps, error) {
	asMaps := &GtmAutonomousSystemMaps{service: domain.parent.service}
	if err := domain.getDomainItems("as-maps", &asMaps); err != nil {
		return nil, err
	}

	return asMaps, nil

}

func (domain *GtmDomain) getDomainItems(itemType string, result interface{}) error {
	res, err := domain.parent.service.client.Get(
		fmt.Sprintf(
			"/config-gtm/v1/domains/%s/%s",
			domain.domain,
			itemType,
		),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	err = res.BodyJson(&result)

	if err != nil {
		return err
	}

	return nil
}
