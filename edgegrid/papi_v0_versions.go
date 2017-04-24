package edgegrid

import (
	"errors"
	"fmt"
	"time"
)

type PapiVersions struct {
	resource
	service      *PapiV0Service
	PropertyID   string `json:"propertyId"`
	PropertyName string `json:"propertyName"`
	AccountID    string `json:"accountId"`
	ContractID   string `json:"contractId"`
	GroupID      string `json:"groupId"`
	Versions     struct {
		Items []*PapiVersion `json:"items"`
	} `json:"versions"`
	RuleFormat string `json:"ruleFormat,omitempty"`
}

func NewPapiVersions(service *PapiV0Service) *PapiVersions {
	version := &PapiVersions{service: service}
	version.Init()

	return version
}

func (versions *PapiVersions) PostUnmarshalJSON() error {
	versions.Init()

	for key, _ := range versions.Versions.Items {
		versions.Versions.Items[key].parent = versions
	}
	versions.Complete <- true

	return nil
}

// GetVersions populates PapiVersions with property version data
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listversions
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/{?contractId,groupId}
func (versions *PapiVersions) GetVersions(property *PapiProperty, contract *PapiContract, group *PapiGroup) error {
	if property == nil {
		return errors.New("You must provide a property")
	}

	if contract == nil {
		contract = property.Contract
	}

	if group == nil {
		group = property.Group
	}

	res, err := versions.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions?contractId=%s&groupId=%s",
			property.PropertyID,
			contract.ContractID,
			group.GroupID,
		),
	)

	if err != nil {
		return err
	}

	newVersions := NewPapiVersions(versions.service)
	if err = res.BodyJSON(newVersions); err != nil {
		return err
	}

	*versions = *newVersions

	return nil
}

// GetLatatestVersion retrieves the latest PapiVersion for a property
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getthelatestversion
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/latest{?contractId,groupId,activatedOn}
// Todo: Mimic behavior of and fallback to /papi/v0/properties/{propertyId}/versions/latest{?contractId,groupId,activatedOn}
// Todo: Move to PapiProperty.GetLatestVersion
func (versions *PapiVersions) GetLatestVersion() *PapiVersion {
	if len(versions.Versions.Items) > 0 {
		return versions.Versions.Items[len(versions.Versions.Items)-1]
	}
	return nil
}

// Todo: refactor to wrap PapiVersion.NewVersion() for createFromVersion behavior
func (versions *PapiVersions) NewVersion(createFromVersion *PapiVersion, useEtagStrict bool) *PapiVersion {
	if createFromVersion == nil {
		createFromVersion = versions.GetLatestVersion()
	}

	version := NewPapiVersion(versions)
	version.CreateFromVersion = createFromVersion.PropertyVersion

	versions.Versions.Items = append(versions.Versions.Items, version)

	if useEtagStrict {
		version.CreateFromVersionEtag = createFromVersion.Etag
	}

	return version
}

type PapiVersion struct {
	resource
	parent                *PapiVersions
	PropertyVersion       int       `json:"propertyVersion,omitempty"`
	UpdatedByUser         string    `json:"updatedByUser,omitempty"`
	UpdatedDate           time.Time `json:"updatedDate,omitempty"`
	ProductionStatus      string    `json:"productionStatus,omitempty"`
	StagingStatus         string    `json:"stagingStatus,omitempty"`
	Etag                  string    `json:"etag,omitempty"`
	ProductID             string    `json:"productId,omitempty"`
	Note                  string    `json:"note,omitempty"`
	CreateFromVersion     int       `json:"createFromVersion,omitempty"`
	CreateFromVersionEtag string    `json:"createFromVersionEtag,omitempty"`
	Complete              chan bool `json:"-"`
}

func NewPapiVersion(parent *PapiVersions) *PapiVersion {
	version := &PapiVersion{parent: parent}
	version.Init()

	return version
}

func (version *PapiVersion) HasBeenActivated() (bool, error) {
	properties := NewPapiProperties(version.parent.service)
	property := NewPapiProperty(properties)
	property.PropertyID = version.parent.PropertyID

	property.Group = NewPapiGroup(NewPapiGroups(version.parent.service))
	property.Group.GroupID = version.parent.GroupID
	go property.Group.GetGroup()

	property.Contract = NewPapiContract(NewPapiContracts(version.parent.service))
	property.Contract.ContractID = version.parent.ContractID
	go property.Contract.GetContract()

	go (func(property *PapiProperty) {
		contractCompleted := <-property.Contract.Complete
		groupCompleted := <-property.Group.Complete
		property.Complete <- (contractCompleted && groupCompleted)
	})(property)

	activations, err := property.GetActivations()
	if err != nil {
		return false, err
	}

	for _, activation := range activations.Activations.Items {
		if activation.PropertyVersion == version.PropertyVersion {
			return true, nil
		}
	}

	return false, nil
}

// Save creates a new version
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#createanewversion
// Endpoint: POST /papi/v0/properties/{propertyId}/versions/{?contractId,groupId}
func (version *PapiVersion) Save() error {
	if version.PropertyVersion != 0 {
		return fmt.Errorf("version (%d) already exists", version.PropertyVersion)
	}

	res, err := version.parent.service.client.PostJSON(
		fmt.Sprintf(
			"/papi/v0/properties/{propertyId}/versions/?contractId=%s&groupId=%s",
			version.parent.ContractID,
			version.parent.ContractID,
		),
		version,
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

	res, err = version.parent.service.client.Get(
		location["versionLink"].(string),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	versions := NewPapiVersions(version.parent.service)
	if err = res.BodyJSON(versions); err != nil {
		return err
	}

	newVersion := versions.Versions.Items[0]
	newVersion.parent = version.parent
	version.parent.Versions.Items = append(version.parent.Versions.Items, versions.Versions.Items...)

	*version = *newVersion

	return nil
}
