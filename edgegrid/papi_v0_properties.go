package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

type PapiProperties struct {
	service    *PapiV0Service
	Properties struct {
		Items []*PapiProperty `json:"items"`
	} `json:"properties"`
	Complete chan bool `json:"-"`
}

func NewPapiProperties(service *PapiV0Service) *PapiProperties {
	properties := &PapiProperties{service: service}
	properties.Init()

	return properties
}

func (properties *PapiProperties) Init() {
	properties.Complete = make(chan bool, 1)
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

func (properties *PapiProperties) NewProperty() PapiProperty {
	return PapiProperty{parent: properties}
}

type PapiProperty struct {
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
	Complete          chan bool              `json:"-"`
}

func NewPapiProperty(parent *PapiProperties) PapiProperty {
	property := PapiProperty{parent: parent}
	property.Init()

	return property
}

func (property *PapiProperty) Init() {
	property.Complete = make(chan bool, 1)
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
	property.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions/%d/hostnames/?contractId=%s&groupId=%s",
			property.PropertyId,
			version,
			property.Contract.ContractId,
			property.Group.GroupId,
		),
	)

	return nil, nil
}

func (property *PapiProperty) PostUnmarshalJSON() error {
	property.Init()

	property.Contract = NewPapiContract(nil)
	property.Contract.ContractId = property.ContractId

	property.Group = NewPapiGroup(nil)
	property.Group.GroupId = property.GroupId

	go (func(property *PapiProperty) {
		groups, err := property.parent.service.GetGroups()
		if err != nil {
			return
		}

		for _, group := range groups.Groups.Items {
			if group.GroupId == property.Group.GroupId {
				property.Group.parent = group.parent
				property.Group.ContractIds = group.ContractIds
				property.Group.GroupName = group.GroupName
				property.Group.ParentGroupId = group.ParentGroupId
				property.Group.Complete <- true
			}
		}
		property.Group.Complete <- false
	})(property)

	go (func(property *PapiProperty) {
		contracts, err := property.parent.service.GetContracts()
		if err != nil {
			return
		}

		for _, contract := range contracts.Contracts.Items {
			if contract.ContractId == property.Contract.ContractId {
				property.Contract.parent = contract.parent
				property.Contract.ContractTypeName = contract.ContractTypeName
				property.Contract.Complete <- true
			}
		}
		property.Contract.Complete <- false
	})(property)

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
	res.BodyJson(properties)

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
	PropertyId           string    `json:"propertyId"`
	Version              int       `json:"version"`
	CopyHostnames        bool      `json:"copyHostnames,omitempty"`
	CloneFromVersionEtag string    `json:"cloneFromVersionEtag,omitempty"`
	Complete             chan bool `json:"-"`
}

func NewPapiClonePropertyFrom() *PapiClonePropertyFrom {
	clonePropertyFrom := &PapiClonePropertyFrom{}
	clonePropertyFrom.Init()

	return clonePropertyFrom
}

func (clonePropertyFrom *PapiClonePropertyFrom) Init() {
	clonePropertyFrom.Complete = make(chan bool, 1)
}

func (clonePropertyFrom *PapiClonePropertyFrom) PostUnmashalJSON() error {
	clonePropertyFrom.Init()
	clonePropertyFrom.Complete <- true

	return nil
}
