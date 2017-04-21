package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

type PapiProperties struct {
	Resource
	service    *PapiV0Service
	Properties struct {
		Items []*PapiProperty `json:"items"`
	} `json:"properties"`
}

func NewPapiProperties(service *PapiV0Service) *PapiProperties {
	properties := &PapiProperties{service: service}
	properties.Init()

	return properties
}

func (properties *PapiProperties) PostUnmarshalJSON() error {
	properties.Init()

	for key, property := range properties.Properties.Items {
		properties.Properties.Items[key].parent = properties
		if property, ok := json.ImplementsPostJsonUnmarshaler(property); ok {
			err := property.(json.PostJsonUnmarshaler).PostUnmarshalJSON()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (properties *PapiProperties) GetProperties(contract *PapiContract, group *PapiGroup) error {
	if contract == nil {
		contract = NewPapiContract(NewPapiContracts(properties.service))
		contract.ContractId = group.ContractIds[0]
	}

	res, err := properties.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties?groupId=%s&contractId=%s",
			group.GroupId,
			contract.ContractId,
		),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	newProperties := NewPapiProperties(properties.service)
	if err = res.BodyJson(newProperties); err != nil {
		return err
	}

	*properties = *newProperties

	return nil
}

func (properties *PapiProperties) AddProperty(newProperty *PapiProperty) {
	if newProperty.Group.GroupId != "" {
		for key, property := range properties.Properties.Items {
			if property.PropertyId == newProperty.PropertyId {
				properties.Properties.Items[key] = newProperty
				return
			}
		}
	}

	newProperty.parent = properties

	properties.Properties.Items = append(properties.Properties.Items, newProperty)
}

func (properties *PapiProperties) FindProperty(name string) (*PapiProperty, error) {
	var property *PapiProperty
	var propertyFound bool
	for _, property = range properties.Properties.Items {
		if property.PropertyName == name {
			propertyFound = true
			break
		}
	}

	if !propertyFound {
		return nil, fmt.Errorf("Unable to find property: \"%s\"", name)
	}

	return property, nil
}

func (properties *PapiProperties) NewProperty(contract *PapiContract, group *PapiGroup) *PapiProperty {
	property := NewPapiProperty(properties)
	property.Contract = contract
	property.Group = group
	go property.Contract.GetContract()
	go property.Group.GetGroup()
	go (func(property *PapiProperty) {
		groupCompleted := <-property.Group.Complete
		contractCompleted := <-property.Contract.Complete
		property.Complete <- (groupCompleted && contractCompleted)
	})(property)

	return property
}

type PapiProperty struct {
	Resource
	parent            *PapiProperties
	AccountId         string                 `json:"accountId,omitempty"`
	Contract          *PapiContract          `json:"-"`
	Group             *PapiGroup             `json:"-"`
	ContractId        string                 `json:"contractId,omitempty"`
	GroupId           string                 `json:"groupId,omitempty"`
	PropertyId        string                 `json:"propertyId,omitempty"`
	PropertyName      string                 `json:"propertyName"`
	LatestVersion     int                    `json:"latestVersion,omitempty"`
	StagingVersion    int                    `json:"stagingVersion,omitempty"`
	ProductionVersion int                    `json:"productionVersion,omitempty"`
	Note              string                 `json:"note,omitempty"`
	ProductId         string                 `json:"productId,omitempty"`
	CloneFrom         *PapiClonePropertyFrom `json:"cloneFrom"`
}

func NewPapiProperty(parent *PapiProperties) *PapiProperty {
	property := &PapiProperty{parent: parent}
	property.Init()
	return property
}

func (property *PapiProperty) PreMarshalJSON() error {
	property.GroupId = property.Group.GroupId
	property.ContractId = property.Contract.ContractId
	return nil
}

func (property *PapiProperty) GetActivations() (*PapiActivations, error) {
	res, err := property.parent.service.client.Get(
		fmt.Sprintf("/papi/v0/properties/%s/activations?contractId=%s&groupId=%s",
			property.PropertyId,
			property.Contract.ContractId,
			property.Group.GroupId,
		),
	)

	if err != nil {
		return nil, err
	}

	activations := &PapiActivations{service: property.parent.service}
	if err = res.BodyJson(activations); err != nil {
		return nil, err
	}

	return activations, nil
}

func (property *PapiProperty) GetAvailableBehaviors() (*PapiAvailableBehaviors, error) {
	// /papi/v0/properties/{propertyId}/versions/{propertyVersion}/available-behaviors{?contractId,groupId}
	res, err := property.parent.service.client.Get(fmt.Sprintf(
		"/papi/v0/properties/%s/versions/%d/available-behaviors?contractId=%s&groupId=%s",
		property.PropertyId,
		property.LatestVersion,
		property.Contract.ContractId,
		property.Group.GroupId,
	))

	if err != nil {
		return nil, err
	}

	behaviors := &PapiAvailableBehaviors{service: property.parent.service}
	if err = res.BodyJson(behaviors); err != nil {
		return nil, err
	}

	return behaviors, nil
}

func (property *PapiProperty) GetRules() (*PapiRules, error) {
	// /papi/v0/properties/{propertyId}/versions/{propertyVersion}/rules/{?contractId,groupId}
	res, err := property.parent.service.client.Get(fmt.Sprintf(
		"/papi/v0/properties/%s/versions/%d/rules?contractId=%s&groupId=%s",
		property.PropertyId,
		property.LatestVersion,
		property.Contract.ContractId,
		property.Group.GroupId,
	))

	if err != nil {
		return nil, err
	}

	rules := &PapiRules{service: property.parent.service}
	if err = res.BodyJson(rules); err != nil {
		return nil, err
	}

	return rules, nil
}

func (property *PapiProperty) GetVersions() (*PapiVersions, error) {
	// /papi/v0/properties/{propertyId}/versions/{?contractId,groupId}
	res, err := property.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions?contractId=%s&groupId=%s",
			property.PropertyId,
			property.Contract.ContractId,
			property.Group.GroupId,
		),
	)

	if err != nil {
		return nil, err
	}

	versions := &PapiVersions{service: property.parent.service}
	if err = res.BodyJson(versions); err != nil {
		return nil, err
	}

	return versions, nil

}

func (property *PapiProperty) GetHostnames(version int) (*PapiHostnames, error) {
	if version == 0 {
		version = property.LatestVersion
	}

	res, err := property.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions/%d/hostnames/?contractId=%s&groupId=%s",
			property.PropertyId,
			version,
			property.Contract.ContractId,
			property.Group.GroupId,
		),
	)

	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, NewApiError(res)
	}

	hostnames := NewPapiHostnames(property.parent.service)
	if err := res.BodyJson(hostnames); err != nil {
		return nil, err
	}

	return hostnames, nil
}

func (property *PapiProperty) PostUnmarshalJSON() error {
	property.Init()

	property.Contract = NewPapiContract(&PapiContracts{service: property.parent.service})
	property.Contract.ContractId = property.ContractId

	property.Group = NewPapiGroup(&PapiGroups{service: property.parent.service})
	property.Group.GroupId = property.GroupId

	go property.Group.GetGroup()
	go property.Contract.GetContract()

	go (func(property *PapiProperty) {
		contractComplete := <-property.Contract.Complete
		groupComplete := <-property.Group.Complete
		property.Complete <- (contractComplete && groupComplete)
	})(property)

	return nil
}

func (property *PapiProperty) Save() error {
	res, err := property.parent.service.client.PostJson(
		fmt.Sprintf(
			"/papi/v0/properties?contractId=%s&groupId=%s",
			property.Contract.ContractId,
			property.Group.GroupId,
		),
		property,
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	var location map[string]interface{}
	res.BodyJson(&location)

	res, err = property.parent.service.client.Get(
		location["propertyLink"].(string),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	properties := &PapiProperties{service: property.parent.service}
	err = res.BodyJson(properties)
	if err != nil {
		return err
	}

	newProperty := properties.Properties.Items[0]
	newProperty.parent = property.parent
	property.parent.Properties.Items = append(property.parent.Properties.Items, newProperty)

	*property = *newProperty

	return nil
}

func (property *PapiProperty) Delete() error {
	// /papi/v0/properties/{propertyId}{?contractId,groupId}
	res, err := property.parent.service.client.Delete(
		fmt.Sprintf(
			"/papi/v0/properties/%s?contractId=%s&groupId=%s",
			property.PropertyId,
			property.Contract.ContractId,
			property.Group.GroupId,
		),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	return nil
}

type PapiClonePropertyFrom struct {
	Resource
	PropertyId           string `json:"propertyId"`
	Version              int    `json:"version"`
	CopyHostnames        bool   `json:"copyHostnames,omitempty"`
	CloneFromVersionEtag string `json:"cloneFromVersionEtag,omitempty"`
}

func NewPapiClonePropertyFrom() *PapiClonePropertyFrom {
	clonePropertyFrom := &PapiClonePropertyFrom{}
	clonePropertyFrom.Init()

	return clonePropertyFrom
}
