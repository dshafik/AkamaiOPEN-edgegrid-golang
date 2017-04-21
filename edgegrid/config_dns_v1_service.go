package edgegrid

import (
	"log"
)

type ConfigDNSV1Service struct {
	client *Client
	config *Config
}

func NewConfigDNSV1Service(client *Client, config *Config) *ConfigDNSV1Service {
	service := ConfigDNSV1Service{client: client, config: config}
	return &service
}

func (service *ConfigDNSV1Service) NewZone(hostname string) DNSZone {
	zone := DNSZone{service: service, Token: "new"}
	zone.Zone.Name = hostname
	return zone
}

func (service *ConfigDNSV1Service) GetZone(hostname string) (*DNSZone, error) {
	zone := service.NewZone(hostname)
	res, err := service.client.Get("/config-dns/v1/zones/" + hostname)
	if err != nil {
		return nil, err
	}

	if res.IsError() && res.StatusCode != 404 {
		return nil, NewAPIError(res)
	} else if res.StatusCode == 404 {
		log.Printf("[DEBUG] Zone \"%s\" not found, creating new zone.", hostname)
		zone = service.NewZone(hostname)
		return &zone, nil
	} else {
		err = res.BodyJSON(&zone)
		if err != nil {
			return nil, err
		}

		zone.marshalRecords()

		return &zone, nil
	}
}
