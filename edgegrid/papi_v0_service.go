package edgegrid

import (
	"errors"
	"fmt"
)

type PapiV0Service struct {
	client *Client
	config *Config
}

func NewPapiV0Service(client *Client, config *Config) PapiV0Service {
	return PapiV0Service{client: client, config: config}
}

func (papi *PapiV0Service) GetGroups() (*PapiGroups, error) {
	res, err := papi.client.Get("/papi/v0/groups")
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, NewApiError(res)
	}

	groups := &PapiGroups{service: papi}
	if err = res.BodyJson(groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (papi *PapiV0Service) GetContracts() (*PapiContracts, error) {
	res, err := papi.client.Get("/papi/v0/contracts")
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, NewApiError(res)
	}

	contracts := &PapiContracts{service: papi}
	if err = res.BodyJson(contracts); err != nil {
		return nil, err
	}

	return contracts, nil
}

func (papi *PapiV0Service) GetProperties(contract *PapiContract, group *PapiGroup) (*PapiProperties, error) {
	if contract == nil {
		contract = NewPapiContract(nil)
		contract.ContractId = group.ContractIds[0]
	}
	res, err := group.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties?groupId=%s&contractId=%s",
			group.GroupId,
			contract.ContractId,
		),
	)
	if err != nil {
		return nil, err
	}

	if res.IsError() == true {
		return nil, NewApiError(res)
	}

	properties := NewPapiProperties(group.parent.service)
	err = res.BodyJson(properties)
	if err != nil {
		return nil, err
	}

	return properties, nil
}

func (papi *PapiV0Service) GetEdgeHostnames(contract *PapiContract, group *PapiGroup, options string) (*PapiEdgeHostnames, error) {
	if contract == nil {
		contract = &PapiContract{ContractId: group.ContractIds[0]}
	}

	if options != "" {
		options = fmt.Sprintf("&options=%s", options)
	}

	res, err := group.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/edgehostnames?groupId=%s&contractId=%s%s",
			group.GroupId,
			contract.ContractId,
			options,
		),
	)
	if err != nil {
		return nil, err
	}

	if res.IsError() == true {
		return nil, NewApiError(res)
	}

	edgeHostnames := &PapiEdgeHostnames{service: group.parent.service}
	err = res.BodyJson(edgeHostnames)
	if err != nil {
		return nil, err
	}

	return edgeHostnames, nil
}

func (papi *PapiV0Service) GetCpCodes(contract *PapiContract, group *PapiGroup) (*PapiCpCodes, error) {
	if contract == nil {
		contract = &PapiContract{ContractId: group.ContractIds[0]}
	}
	res, err := group.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/cpcodes?groupId=%s&contractId=%s",
			group.GroupId,
			contract.ContractId,
		),
	)
	if err != nil {
		return nil, err
	}

	if res.IsError() == true {
		return nil, NewApiError(res)
	}

	cpcodes := &PapiCpCodes{service: group.parent.service}
	err = res.BodyJson(cpcodes)
	if err != nil {
		return nil, err
	}

	return cpcodes, nil
}

func (papi *PapiV0Service) GetVersions(property *PapiProperty, contract *PapiContract, group *PapiGroup) (*PapiVersions, error) {
	// /papi/v0/properties/{propertyId}/versions/{?contractId,groupId}
	if property == nil {
		return nil, errors.New("You must provide a property")
	}

	if contract == nil {
		contract = property.Contract
	}

	if group == nil {
		group = property.Group
	}

	res, err := property.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions?contractId=%s&groupId=%s",
			property.PropertyId,
			contract.ContractId,
			group.GroupId,
		),
	)

	if err != nil {
		return nil, err
	}

	versions := &PapiVersions{service: papi}
	if err = res.BodyJson(versions); err != nil {
		return nil, err
	}

	return versions, nil
}

func (papi *PapiV0Service) NewProperty(contract *PapiContract, group *PapiGroup) (*PapiProperty, error) {
	if contract == nil {
		contract = &PapiContract{ContractId: group.ContractIds[0]}
	}

	properties := &PapiProperties{service: group.parent.service}
	property := properties.NewProperty(contract, group)

	return property, nil
}
