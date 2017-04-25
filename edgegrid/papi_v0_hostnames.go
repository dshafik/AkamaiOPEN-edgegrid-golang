package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

// PapiHostnames is a collection of Property Hostnames
type PapiHostnames struct {
	resource
	service         *PapiV0Service
	AccountID       string `json:"accountId"`
	ContractID      string `json:"contractId"`
	GroupID         string `json:"groupId"`
	PropertyID      string `json:"propertyId"`
	PropertyVersion int    `json:"propertyVersion"`
	Etag            string `json:"etag"`
	Hostnames       struct {
		Items []*PapiHostname `json:"items"`
	} `json:"hostnames"`
}

// NewPapiHostnames creates a new PapiHostnames
func NewPapiHostnames(service *PapiV0Service) *PapiHostnames {
	hostnames := &PapiHostnames{service: service}
	hostnames.Init()

	return hostnames
}

// PostUnmarshalJSON is called after JSON unmarshaling into PapiEdgeHostnames
//
// See: edgegrid/json.Unmarshal()
func (hostnames *PapiHostnames) PostUnmarshalJSON() error {
	hostnames.Init()

	for key, hostname := range hostnames.Hostnames.Items {
		hostnames.Hostnames.Items[key].parent = hostnames
		if hostname, ok := json.ImplementsPostJSONUnmarshaler(hostname); ok {
			if err := hostname.(json.PostJSONUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}

	hostnames.Complete <- true

	return nil
}

// GetHostnames retrieves hostnames assigned to a given property
//
// If no version is given, the latest version is used
//
// See: PapiProperty.GetHostnames()
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listapropertyshostnames
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/{propertyVersion}/hostnames/{?contractId,groupId}
func (hostnames *PapiHostnames) GetHostnames(version *PapiVersion) error {
	if version == nil {
		property := NewPapiProperty(NewPapiProperties(hostnames.service))
		property.PropertyID = hostnames.PropertyID
		err := property.GetProperty()
		if err != nil {
			return err
		}

		version, err = property.GetLatestVersion("")
		if err != nil {
			return err
		}
	}

	res, err := hostnames.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions/%d/hostnames/?contractId=%s&groupId=%s",
			hostnames.PropertyID,
			version.PropertyVersion,
			hostnames.ContractID,
			hostnames.GroupID,
		),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newHostnames := NewPapiHostnames(hostnames.service)
	if err = res.BodyJSON(newHostnames); err != nil {
		return err
	}

	*hostnames = *newHostnames

	return nil
}

// Save updates a properties hostnames
func (hostnames *PapiHostnames) Save() error {
	req, err := hostnames.service.client.NewJSONRequest(
		"PUT",
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions/%d/hostnames?contractId=%s&groupId%s",
			hostnames.PropertyID,
			hostnames.PropertyVersion,
			hostnames.ContractID,
			hostnames.GroupID,
		),
		hostnames.Hostnames.Items,
	)

	if err != nil {
		return err
	}

	req.Header.Set("If-Match", hostnames.Etag)

	res, err := hostnames.service.client.Do(req)
	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newHostnames := NewPapiHostnames(hostnames.service)
	if err = res.BodyJSON(newHostnames); err != nil {
		return err
	}

	*hostnames = *newHostnames

	return nil
}

// PapiHostname represents a property hostname resource
type PapiHostname struct {
	resource
	parent         *PapiHostnames
	CnameType      PapiCnameTypeValue `json:"cnameType"`
	EdgeHostnameID string             `json:"edgeHostnameId"`
	CnameFrom      string             `json:"cnameFrom"`
	CnameTo        string             `json:"cnameTo,omitempty"`
}

// NewPapiHostname creates a new PapiHostname
func NewPapiHostname(parent *PapiHostnames) *PapiHostname {
	hostname := &PapiHostname{parent: parent, CnameType: PapiCnameTypeEdgeHostname}
	hostname.Init()

	return hostname
}

// PapiCnameTypeValue is used to create an "enum" of possible PapiHostname.CnameType values
type PapiCnameTypeValue string

const (
	// PapiCnameTypeEdgeHostname PapiHostname.CnameType value EDGE_HOSTNAME
	PapiCnameTypeEdgeHostname PapiCnameTypeValue = "EDGE_HOSTNAME"
)
