package edgegrid

import (
	"errors"
	"fmt"
	"time"
)

// PapiVersions contains a collection of Property Versions
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

// NewPapiVersions creates a new PapiVersions
func NewPapiVersions(service *PapiV0Service) *PapiVersions {
	version := &PapiVersions{service: service}
	version.Init()

	return version
}

// PostUnmarshalJSON is called after JSON unmarshaling into PapiEdgeHostnames
//
// See: edgegrid/json.Unmarshal()
func (versions *PapiVersions) PostUnmarshalJSON() error {
	versions.Init()

	for key := range versions.Versions.Items {
		versions.Versions.Items[key].parent = versions
	}
	versions.Complete <- true

	return nil
}

// AddVersion adds or replaces a version within the collection
func (versions *PapiVersions) AddVersion(version *PapiVersion) {
	if version.PropertyVersion != 0 {
		for key, v := range versions.Versions.Items {
			if v.PropertyVersion == version.PropertyVersion {
				versions.Versions.Items[key] = version
				return
			}
		}
	}

	versions.Versions.Items = append(versions.Versions.Items, version)
}

// GetVersions retrieves all versions for a a given property
//
// See: PapiProperty.GetVersions()
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listversions
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/{?contractId,groupId}
func (versions *PapiVersions) GetVersions(property *PapiProperty) error {
	if property == nil {
		return errors.New("You must provide a property")
	}

	res, err := versions.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions?contractId=%s&groupId=%s",
			property.PropertyID,
			property.Contract.ContractID,
			property.Group.GroupID,
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

// GetLatestVersion retrieves the latest PapiVersion for a property
//
// See: PapiProperty.GetLatestVersion()
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getthelatestversion
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/latest{?contractId,groupId,activatedOn}
func (versions *PapiVersions) GetLatestVersion(activatedOn PapiNetworkValue) (*PapiVersion, error) {
	if activatedOn != "" {
		activatedOn = "&activatedOn=" + activatedOn
	}

	res, err := versions.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions/latest?contractId=%s&groupId=%s%s",
			versions.PropertyID,
			versions.ContractID,
			versions.GroupID,
			activatedOn,
		),
	)

	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, NewAPIError(res)
	}

	latest := NewPapiVersion(versions)
	if err := res.BodyJSON(latest); err != nil {
		return nil, err
	}

	return latest, nil
}

// NewVersion creates a new version associated with the PapiVersions collection
func (versions *PapiVersions) NewVersion(createFromVersion *PapiVersion, useEtagStrict bool) *PapiVersion {
	if createFromVersion == nil {
		var err error
		createFromVersion, err = versions.GetLatestVersion("")
		if err != nil {
			return nil
		}
	}

	version := NewPapiVersion(versions)
	version.CreateFromVersion = createFromVersion.PropertyVersion

	versions.Versions.Items = append(versions.Versions.Items, version)

	if useEtagStrict {
		version.CreateFromVersionEtag = createFromVersion.Etag
	}

	return version
}

// PapiVersion represents a Property Version
type PapiVersion struct {
	resource
	parent                *PapiVersions
	PropertyVersion       int             `json:"propertyVersion,omitempty"`
	UpdatedByUser         string          `json:"updatedByUser,omitempty"`
	UpdatedDate           time.Time       `json:"updatedDate,omitempty"`
	ProductionStatus      PapiStatusValue `json:"productionStatus,omitempty"`
	StagingStatus         PapiStatusValue `json:"stagingStatus,omitempty"`
	Etag                  string          `json:"etag,omitempty"`
	ProductID             string          `json:"productId,omitempty"`
	Note                  string          `json:"note,omitempty"`
	CreateFromVersion     int             `json:"createFromVersion,omitempty"`
	CreateFromVersionEtag string          `json:"createFromVersionEtag,omitempty"`
	Complete              chan bool       `json:"-"`
}

// NewPapiVersion creates a new PapiVersion
func NewPapiVersion(parent *PapiVersions) *PapiVersion {
	version := &PapiVersion{parent: parent}
	version.Init()

	return version
}

// GetVersion populates a PapiVersion
//
// Api Docs: https://developer.akamai.com/api/luna/papi/resources.html#getaversion
// Endpoint: /papi/v0/properties/{propertyId}/versions/{propertyVersion}{?contractId,groupId}
func (version *PapiVersion) GetVersion(property *PapiProperty, getVersion int) error {
	if getVersion == 0 {
		getVersion = property.LatestVersion
	}

	res, err := version.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions/%d?contractId=%s&groupId=%s",
			property.PropertyID,
			getVersion,
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

	newVersion := NewPapiVersion(version.parent)
	if err := res.BodyJSON(newVersion); err != nil {
		return err
	}

	*version = *newVersion

	return nil
}

// HasBeenActivated determines if a given version has been activated, optionally on a specific network
func (version *PapiVersion) HasBeenActivated(activatedOn PapiNetworkValue) (bool, error) {
	properties := NewPapiProperties(version.parent.service)
	property := NewPapiProperty(properties)
	property.PropertyID = version.parent.PropertyID

	property.Group = NewPapiGroup(NewPapiGroups(version.parent.service))
	property.Group.GroupID = version.parent.GroupID

	property.Contract = NewPapiContract(NewPapiContracts(version.parent.service))
	property.Contract.ContractID = version.parent.ContractID

	activations, err := property.GetActivations()
	if err != nil {
		return false, err
	}

	for _, activation := range activations.Activations.Items {
		if activation.PropertyVersion == version.PropertyVersion && (activatedOn == "" || activation.Network == activatedOn) {
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
			"/papi/v0/properties/%s/versions/?contractId=%s&groupId=%s",
			version.parent.PropertyID,
			version.parent.ContractID,
			version.parent.GroupID,
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
