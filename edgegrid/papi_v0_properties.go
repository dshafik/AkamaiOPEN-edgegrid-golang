package edgegrid

import (
	"encoding/json"
	"errors"
	"fmt"
)

type PapiProperties struct {
	service    *PapiV0Service
	Properties struct {
		Items []*PapiProperty `json:"items"`
	} `json:"properties"`
}

func (properties *PapiProperties) UnmarshalJSON(b []byte) error {
	type PapiPropertiesTemp PapiProperties
	temp := &PapiPropertiesTemp{service: properties.service}

	if err := json.Unmarshal(b, temp); err != nil {
		return err
	}
	*properties = (PapiProperties)(*temp)

	for key, _ := range properties.Properties.Items {
		properties.Properties.Items[key].parent = properties
	}

	return nil
}

func (properties *PapiProperties) AddProperty(newProperty *PapiProperty) {
	if newProperty.GroupId != "" {
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
		return nil, errors.New(fmt.Sprintf("Unable to find property: \"%s\"", name))
	}

	return property, nil
}

func (properties *PapiProperties) NewProperty() *PapiProperty {
	return &PapiProperty{parent: properties}
}

type PapiProperty struct {
	parent            *PapiProperties
	AccountId         string                 `json:"accountId,omitempty"`
	Contract          *PapiContract          `json:"-"`
	ContractId        string                 `json:"contractId,omitempty"`
	Group             *PapiGroup             `json:"-"`
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

func (property *PapiProperty) UnmarshalJSON(b []byte) error {
	type PapiPropertyTemp PapiProperty
	temp := &PapiPropertyTemp{}

	if err := json.Unmarshal(b, temp); err != nil {
		return err
	}
	*property = (PapiProperty)(*temp)

	property.Contract = &PapiContract{
		ContractId: property.ContractId,
	}
	property.Group = &PapiGroup{
		GroupId: property.GroupId,
	}

	return nil
}

func (property *PapiProperty) MarshalJSON() ([]byte, error) {
	type PapiPropertyTemp PapiProperty
	temp := (PapiPropertyTemp)(*property)
	temp.ContractId = ""
	temp.GroupId = ""

	return json.Marshal(temp)
}

func (property *PapiProperty) GetActivations() (*PapiActivations, error) {
	res, err := property.parent.service.client.Get(
		fmt.Sprintf("/papi/v0/properties/%s/activations?contractId=%s&groupId=%s",
			property.PropertyId,
			property.ContractId,
			property.GroupId,
		),
	)

	if err != nil {
		return nil, err
	}

	activations := &PapiActivations{service: property.parent.service}
	if err = res.BodyJson(&activations); err != nil {
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
		property.ContractId,
		property.GroupId,
	))

	if err != nil {
		return nil, err
	}

	behaviors := &PapiAvailableBehaviors{service: property.parent.service}
	if err = res.BodyJson(&behaviors); err != nil {
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
		property.ContractId,
		property.GroupId,
	))

	if err != nil {
		return nil, err
	}

	rules := &PapiRules{service: property.parent.service}
	if err = res.BodyJson(&rules); err != nil {
		return nil, err
	}

	return rules, nil
}

func (property *PapiProperty) GetVersions() (*PapiVersions, error) {
	// /papi/v0/properties/{propertyId}/versions/{?contractId,groupId}
	res, err := property.parent.service.client.Get(fmt.Sprintf(
		"/papi/v0/properties/%s/versions?contractId=%s&groupId=%s",
		property.PropertyId,
		property.ContractId,
		property.GroupId,
	))

	if err != nil {
		return nil, err
	}

	versions := &PapiVersions{service: property.parent.service}
	if err = res.BodyJson(&versions); err != nil {
		return nil, err
	}

	return versions, nil

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
	res.BodyJson(&properties)

	newProperty := properties.Properties.Items[0]
	newProperty.parent = property.parent
	property.parent.Properties.Items = append(property.parent.Properties.Items, properties.Properties.Items...)

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
	PropertyId           string `json:"propertyId"`
	Version              int    `json:"version"`
	CopyHostnames        bool   `json:"copyHostnames,omitempty"`
	CloneFromVersionEtag string `json:"cloneFromVersionEtag,omitempty"`
}
