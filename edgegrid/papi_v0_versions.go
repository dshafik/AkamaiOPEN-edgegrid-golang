package edgegrid

import (
	"fmt"
	"time"
)

type PapiVersions struct {
	Resource
	service      *PapiV0Service
	PropertyId   string `json:"propertyId"`
	PropertyName string `json:"propertyName"`
	AccountId    string `json:"accountId"`
	ContractId   string `json:"contractId"`
	GroupId      string `json:"groupId"`
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
			property.PropertyId,
			contract.ContractId,
			group.GroupId,
		),
	)

	if err != nil {
		return err
	}

	newVersions := NewPapiVersions(versions.service)
	if err = res.BodyJson(versions); err != nil {
		return err
	}

	*versions = *newVersions

	return nil
}

// Todo: Mimic behavior of and fallback to /papi/v0/properties/{propertyId}/versions/latest{?contractId,groupId,activatedOn}
func (versions *PapiVersions) GetLatestVersion() *PapiVersion {
	if len(versions.Versions.Items) > 0 {
		return versions.Versions.Items[len(versions.Versions.Items)-1]
	}
	return nil
}

func (versions *PapiVersions) NewVersion(createFromVersion *PapiVersion, useEtagStrict bool) *PapiVersion {
	if createFromVersion == nil {
		createFromVersion = versions.GetLatestVersion()
	}

	version := NewPapiVersion(versions)
	version.CreateFromVersion = createFromVersion.PropertyVersion

	if useEtagStrict {
		version.CreateFromVersionEtag = createFromVersion.Etag
	}

	return version
}

type PapiVersion struct {
	Resource
	parent                *PapiVersions
	PropertyVersion       int       `json:"propertyVersion,omitempty"`
	UpdatedByUser         string    `json:"updatedByUser,omitempty"`
	UpdatedDate           time.Time `json:"updatedDate,omitempty"`
	ProductionStatus      string    `json:"productionStatus,omitempty"`
	StagingStatus         string    `json:"stagingStatus,omitempty"`
	Etag                  string    `json:"etag,omitempty"`
	ProductId             string    `json:"productId,omitempty"`
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
	property.PropertyId = version.parent.PropertyId

	property.Group = NewPapiGroup(NewPapiGroups(version.parent.service))
	property.Group.GroupId = version.parent.GroupId
	go property.Group.GetGroup()

	property.Contract = NewPapiContract(NewPapiContracts(version.parent.service))
	property.Contract.ContractId = version.parent.ContractId
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

func (version *PapiVersion) Save() error {
	if version.PropertyVersion != 0 {
		return fmt.Errorf("Version (%d) already exists!", version.PropertyVersion)
	}

	res, err := version.parent.service.client.PostJson(
		fmt.Sprintf(
			"/papi/v0/properties/{propertyId}/versions/?contractId=%s&groupId=%s",
			version.parent.ContractId,
			version.parent.ContractId,
		),
		version,
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	var location map[string]interface{}
	res.BodyJson(&location)

	res, err = version.parent.service.client.Get(
		location["versionLink"].(string),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	versions := NewPapiVersions(version.parent.service)
	if err := res.BodyJson(versions); err != nil {
		return err
	}

	newVersion := versions.Versions.Items[0]
	newVersion.parent = version.parent
	version.parent.Versions.Items = append(version.parent.Versions.Items, versions.Versions.Items...)

	*version = *newVersion

	return nil
}
