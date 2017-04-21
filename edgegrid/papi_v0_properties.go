package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

type PapiProperties struct {
	resource
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
		if property, ok := json.ImplementsPostJSONUnmarshaler(property); ok {
			err := property.(json.PostJSONUnmarshaler).PostUnmarshalJSON()
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
		contract.ContractID = group.ContractIDs[0]
	}

	res, err := properties.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties?groupId=%s&contractId=%s",
			group.GroupID,
			contract.ContractID,
		),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newProperties := NewPapiProperties(properties.service)
	if err = res.BodyJSON(newProperties); err != nil {
		return err
	}

	*properties = *newProperties

	return nil
}

func (properties *PapiProperties) AddProperty(newProperty *PapiProperty) {
	if newProperty.Group.GroupID != "" {
		for key, property := range properties.Properties.Items {
			if property.PropertyID == newProperty.PropertyID {
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
	resource
	parent            *PapiProperties
	AccountID         string                 `json:"accountId,omitempty"`
	Contract          *PapiContract          `json:"-"`
	Group             *PapiGroup             `json:"-"`
	ContractID        string                 `json:"contractId,omitempty"`
	GroupID           string                 `json:"groupId,omitempty"`
	PropertyID        string                 `json:"propertyId,omitempty"`
	PropertyName      string                 `json:"propertyName"`
	LatestVersion     int                    `json:"latestVersion,omitempty"`
	StagingVersion    int                    `json:"stagingVersion,omitempty"`
	ProductionVersion int                    `json:"productionVersion,omitempty"`
	Note              string                 `json:"note,omitempty"`
	ProductID         string                 `json:"productId,omitempty"`
	CloneFrom         *PapiClonePropertyFrom `json:"cloneFrom"`
}

func NewPapiProperty(parent *PapiProperties) *PapiProperty {
	property := &PapiProperty{parent: parent}
	property.Init()
	return property
}

func (property *PapiProperty) PreMarshalJSON() error {
	property.GroupID = property.Group.GroupID
	property.ContractID = property.Contract.ContractID
	return nil
}

func (property *PapiProperty) GetActivations() (*PapiActivations, error) {
	res, err := property.parent.service.client.Get(
		fmt.Sprintf("/papi/v0/properties/%s/activations?contractId=%s&groupId=%s",
			property.PropertyID,
			property.Contract.ContractID,
			property.Group.GroupID,
		),
	)

	if err != nil {
		return nil, err
	}

	activations := NewPapiActivations(property.parent.service)
	if err = res.BodyJSON(activations); err != nil {
		return nil, err
	}

	return activations, nil
}

func (property *PapiProperty) GetAvailableBehaviors() (*PapiAvailableBehaviors, error) {
	// /papi/v0/properties/{propertyId}/versions/{propertyVersion}/available-behaviors{?contractId,groupId}
	res, err := property.parent.service.client.Get(fmt.Sprintf(
		"/papi/v0/properties/%s/versions/%d/available-behaviors?contractId=%s&groupId=%s",
		property.PropertyID,
		property.LatestVersion,
		property.Contract.ContractID,
		property.Group.GroupID,
	))

	if err != nil {
		return nil, err
	}

	behaviors := NewPapiAvailableBehaviors(property.parent.service)
	if err = res.BodyJSON(behaviors); err != nil {
		return nil, err
	}

	return behaviors, nil
}

func (property *PapiProperty) GetRules() (*PapiRules, error) {
	// /papi/v0/properties/{propertyId}/versions/{propertyVersion}/rules/{?contractId,groupId}
	res, err := property.parent.service.client.Get(fmt.Sprintf(
		"/papi/v0/properties/%s/versions/%d/rules?contractId=%s&groupId=%s",
		property.PropertyID,
		property.LatestVersion,
		property.Contract.ContractID,
		property.Group.GroupID,
	))

	if err != nil {
		return nil, err
	}

	rules := NewPapiRules(property.parent.service)
	if err = res.BodyJSON(rules); err != nil {
		return nil, err
	}

	return rules, nil
}

func (property *PapiProperty) GetVersions() (*PapiVersions, error) {
	// /papi/v0/properties/{propertyId}/versions/{?contractId,groupId}
	res, err := property.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions?contractId=%s&groupId=%s",
			property.PropertyID,
			property.Contract.ContractID,
			property.Group.GroupID,
		),
	)

	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, NewAPIError(res)
	}

	versions := NewPapiVersions(property.parent.service)
	if err = res.BodyJSON(versions); err != nil {
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
			property.PropertyID,
			version,
			property.Contract.ContractID,
			property.Group.GroupID,
		),
	)

	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, NewAPIError(res)
	}

	hostnames := NewPapiHostnames(property.parent.service)
	if err = res.BodyJSON(hostnames); err != nil {
		return nil, err
	}

	return hostnames, nil
}

func (property *PapiProperty) PostUnmarshalJSON() error {
	property.Init()

	property.Contract = NewPapiContract(NewPapiContracts(property.parent.service))
	property.Contract.ContractID = property.ContractID

	property.Group = NewPapiGroup(NewPapiGroups(property.parent.service))
	property.Group.GroupID = property.GroupID

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
	res, err := property.parent.service.client.PostJSON(
		fmt.Sprintf(
			"/papi/v0/properties?contractId=%s&groupId=%s",
			property.Contract.ContractID,
			property.Group.GroupID,
		),
		property,
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	var location map[string]interface{}
	if err = res.BodyJSON(&location); err != nil {
		return err
	}

	res, err = property.parent.service.client.Get(
		location["propertyLink"].(string),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	properties := NewPapiProperties(property.parent.service)
	if err = res.BodyJSON(properties); err != nil {
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
			property.PropertyID,
			property.Contract.ContractID,
			property.Group.GroupID,
		),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	return nil
}

type PapiClonePropertyFrom struct {
	resource
	PropertyID           string `json:"propertyId"`
	Version              int    `json:"version"`
	CopyHostnames        bool   `json:"copyHostnames,omitempty"`
	CloneFromVersionEtag string `json:"cloneFromVersionEtag,omitempty"`
}

func NewPapiClonePropertyFrom() *PapiClonePropertyFrom {
	clonePropertyFrom := &PapiClonePropertyFrom{}
	clonePropertyFrom.Init()

	return clonePropertyFrom
}
