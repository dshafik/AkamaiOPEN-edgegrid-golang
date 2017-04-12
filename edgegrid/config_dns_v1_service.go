package edgegrid

import (
	"log"
)

type ConfigDnsV1Service struct {
	client *Client
	config *Config
}

func NewConfigDnsV1Service(client *Client, config *Config) *ConfigDnsV1Service {
	service := ConfigDnsV1Service{client: client, config: config}
	return &service
}

func (service *ConfigDnsV1Service) NewZone(hostname string) DnsZone {
	zone := DnsZone{service: service, Token: "new"}
	zone.Zone.Name = hostname
	return zone
}

func (service *ConfigDnsV1Service) GetZone(hostname string) (*DnsZone, error) {
	zone := service.NewZone(hostname)
	res, err := service.client.Get("/config-dns/v1/zones/" + hostname)
	if err != nil {
		return nil, err
	}

	if res.IsError() == true && res.StatusCode != 404 {
		return nil, NewApiError(res)
	} else if res.StatusCode == 404 {
		log.Printf("[DEBUG] Zone \"%s\" not found, creating new zone.", hostname)
		zone = service.NewZone(hostname)
		return &zone, nil
	} else {
		err = res.BodyJson(&zone)
		if err != nil {
			return nil, err
		}

		zone.marshalRecords()

		return &zone, nil
	}
}
