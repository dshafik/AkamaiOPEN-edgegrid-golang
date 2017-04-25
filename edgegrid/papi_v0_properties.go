package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

// PapiProperties is a collection of PAPI Property resources
type PapiProperties struct {
	resource
	service    *PapiV0Service
	Properties struct {
		Items []*PapiProperty `json:"items"`
	} `json:"properties"`
}

// NewPapiProperties creates a new PapiProperties
func NewPapiProperties(service *PapiV0Service) *PapiProperties {
	properties := &PapiProperties{service: service}
	properties.Init()

	return properties
}

// PostUnmarshalJSON is called after JSON unmarshaling into PapiEdgeHostnames
//
// See: edgegrid/json.Unmarshal()
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

// GetProperties populates PapiProperties with property data
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listproperties
// Endpoint: GET /papi/v0/properties/{?contractId,groupId}
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

// AddProperty adds a property to the collection, if the property already exists
// in the collection it will be replaced.
func (properties *PapiProperties) AddProperty(newProperty *PapiProperty) {
	if newProperty.PropertyID != "" {
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

// FindProperty finds a property by name within the collection
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

// NewProperty creates a new property associated with the collection
func (properties *PapiProperties) NewProperty(contract *PapiContract, group *PapiGroup) *PapiProperty {
	property := NewPapiProperty(properties)

	properties.AddProperty(property)

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

// PapiProperty represents a PAPI Property
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

// NewPapiProperty creates a new PapiProperty
func NewPapiProperty(parent *PapiProperties) *PapiProperty {
	property := &PapiProperty{parent: parent}
	property.Init()
	return property
}

// PreMarshalJSON is called before JSON marshaling
//
// See: edgegrid/json.Marshal()
func (property *PapiProperty) PreMarshalJSON() error {
	property.GroupID = property.Group.GroupID
	property.ContractID = property.Contract.ContractID
	return nil
}

// GetProperty populates a PapiProperty
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getaproperty
// Endpoint: GET /papi/v0/properties/{propertyId}{?contractId,groupId}
func (property *PapiProperty) GetProperty() error {
	res, err := property.parent.service.client.Get(
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

	newProperty := NewPapiProperty(property.parent)
	if err := res.BodyJSON(newProperty); err != nil {
		return err
	}

	*property = *newProperty

	return nil
}

// GetActivations retrieves activation data for a given property
//
// See: PapiActivations.GetActivations()
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listactivations
// Endpoint: GET /papi/v0/properties/{propertyId}/activations/{?contractId,groupId}
func (property *PapiProperty) GetActivations() (*PapiActivations, error) {
	activations := NewPapiActivations(property.parent.service)

	if err := activations.GetActivations(property); err != nil {
		return nil, err
	}

	return activations, nil
}

// GetAvailableBehaviors retrieves available behaviors for a given property
//
// See: PapiAvailableBehaviors.GetAvailableBehaviors
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listavailablebehaviors
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/{propertyVersion}/available-behaviors{?contractId,groupId}
func (property *PapiProperty) GetAvailableBehaviors() (*PapiAvailableBehaviors, error) {
	behaviors := NewPapiAvailableBehaviors(property.parent.service)
	if err := behaviors.GetAvailableBehaviors(property); err != nil {
		return nil, err
	}

	return behaviors, nil
}

// GetRules retrieves rules for a property
//
// See: PapiRules.GetRules
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getaruletree
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/{propertyVersion}/rules/{?contractId,groupId}
func (property *PapiProperty) GetRules() (*PapiRules, error) {
	rules := NewPapiRules(property.parent.service)

	if err := rules.GetRules(property); err != nil {
		return nil, err
	}

	return rules, nil
}

// GetVersions retrieves all versions for a a given property
//
// See: PapiVersions.GetVersions()
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listversions
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/{?contractId,groupId}
func (property *PapiProperty) GetVersions() (*PapiVersions, error) {
	versions := NewPapiVersions(property.parent.service)
	err := versions.GetVersions(property)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

// GetLatestVersion gets the latest active version, optionally of a given network
//
// See: PapiVersions.GetLatestVersion()
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getthelatestversion
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/latest{?contractId,groupId,activatedOn}
func (property *PapiProperty) GetLatestVersion(activatedOn PapiNetworkValue) (*PapiVersion, error) {
	versions, err := property.GetVersions()
	if err != nil {
		return nil, err
	}

	return versions.GetLatestVersion(activatedOn)
}

// GetHostnames retrieves hostnames assigned to a given property
//
// If no version is given, the latest version is used
//
// See: PapiHostnames.GetHostnames()
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listapropertyshostnames
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/{propertyVersion}/hostnames/{?contractId,groupId}
func (property *PapiProperty) GetHostnames(version *PapiVersion) (*PapiHostnames, error) {
	hostnames := NewPapiHostnames(property.parent.service)
	hostnames.PropertyID = property.PropertyID
	hostnames.ContractID = property.Contract.ContractID
	hostnames.GroupID = property.Group.GroupID

	if version == nil {
		var err error
		version, err = property.GetLatestVersion("")
		if err != nil {
			return nil, err
		}
	}
	err := hostnames.GetHostnames(version)
	if err != nil {
		return nil, err
	}

	return hostnames, nil
}

// PostUnmarshalJSON is called after JSON unmarshaling into PapiEdgeHostnames
//
// See: edgegrid/json.Unmarshal()
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

// Save will create a property, optionally cloned from another property
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#createorcloneaproperty
// Endpoint: POST /papi/v0/properties/{?contractId,groupId}
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

	var location JSONBody
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

	*property = *newProperty

	return nil
}

// Delete a property
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#removeaproperty
// Endpoint: DELETE /papi/v0/properties/{propertyId}{?contractId,groupId}
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

// PapiClonePropertyFrom represents
type PapiClonePropertyFrom struct {
	resource
	PropertyID           string `json:"propertyId"`
	Version              int    `json:"version"`
	CopyHostnames        bool   `json:"copyHostnames,omitempty"`
	CloneFromVersionEtag string `json:"cloneFromVersionEtag,omitempty"`
}

// NewPapiClonePropertyFrom creates a new PapiClonePropertyFrom
func NewPapiClonePropertyFrom() *PapiClonePropertyFrom {
	clonePropertyFrom := &PapiClonePropertyFrom{}
	clonePropertyFrom.Init()

	return clonePropertyFrom
}
