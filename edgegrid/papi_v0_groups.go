package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

// PapiGroups represents a collection of PAPI groups
type PapiGroups struct {
	resource
	service     *PapiV0Service
	AccountID   string `json:"accountId"`
	AccountName string `json:"accountName"`
	Groups      struct {
		Items []*PapiGroup `json:"items"`
	} `json:"groups"`
}

// NewPapiGroups creates a new PapiGroups
func NewPapiGroups(service *PapiV0Service) *PapiGroups {
	groups := &PapiGroups{service: service}
	return groups
}

// PostUnmarshalJSON is called after JSON unmarshaling into PapiEdgeHostnames
//
// See: edgegrid/json.Unmarshal()
func (groups *PapiGroups) PostUnmarshalJSON() error {
	groups.Init()
	for key, group := range groups.Groups.Items {
		groups.Groups.Items[key].parent = groups
		if group, ok := json.ImplementsPostJSONUnmarshaler(group); ok {
			if err := group.(json.PostJSONUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}

	groups.Complete <- true

	return nil
}

// GetGroups populates PapiGroups with group data
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listgroups
// Endpoint: GET /papi/v0/groups/
func (groups *PapiGroups) GetGroups() error {
	res, err := groups.service.client.Get("/papi/v0/groups")
	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newGroups := NewPapiGroups(groups.service)
	if err = res.BodyJSON(newGroups); err != nil {
		return err
	}

	*groups = *newGroups

	return nil
}

// AddGroup adds a group to a PapiGroups collection
func (groups *PapiGroups) AddGroup(newGroup *PapiGroup) {
	if newGroup.GroupID != "" {
		for key, group := range groups.Groups.Items {
			if group.GroupID == newGroup.GroupID {
				groups.Groups.Items[key] = newGroup
				return
			}
		}
	}

	newGroup.parent = groups

	groups.Groups.Items = append(groups.Groups.Items, newGroup)
}

// FindGroup finds a specific group by name
func (groups *PapiGroups) FindGroup(name string) (*PapiGroup, error) {
	var group *PapiGroup
	var groupFound bool
	for _, group = range groups.Groups.Items {
		if group.GroupName == name {
			groupFound = true
			break
		}
	}

	if !groupFound {
		return nil, fmt.Errorf("Unable to find group: \"%s\"", name)
	}

	return group, nil
}

// PapiGroup represents a group resource
type PapiGroup struct {
	resource
	parent        *PapiGroups
	GroupName     string   `json:"groupName"`
	GroupID       string   `json:"groupId"`
	ParentGroupID string   `json:"parentGroupId,omitempty"`
	ContractIDs   []string `json:"contractIds"`
}

// NewPapiGroup creates a new PapiGroup
func NewPapiGroup(parent *PapiGroups) *PapiGroup {
	group := &PapiGroup{
		parent: parent,
	}
	group.Init()
	return group
}

// GetGroup populates a PapiGroup
func (group *PapiGroup) GetGroup() {
	groups, err := group.parent.service.GetGroups()
	if err != nil {
		return
	}

	for _, g := range groups.Groups.Items {
		if g.GroupID == group.GroupID {
			group.parent = groups
			group.ContractIDs = g.ContractIDs
			group.GroupName = g.GroupName
			group.ParentGroupID = g.ParentGroupID
			group.Complete <- true
			return
		}
	}

	group.Complete <- false
}

// GetProperties retrieves all properties associated with a given group and contract
func (group *PapiGroup) GetProperties(contract *PapiContract) (*PapiProperties, error) {
	return group.parent.service.GetProperties(contract, group)
}

// GetCpCodes retrieves all CP codes associated with a given group and contract
func (group *PapiGroup) GetCpCodes(contract *PapiContract) (*PapiCpCodes, error) {
	return group.parent.service.GetCpCodes(contract, group)
}

// GetEdgeHostnames retrieves all Edge hostnames associated with a given group/contract
func (group *PapiGroup) GetEdgeHostnames(contract *PapiContract, options string) (*PapiEdgeHostnames, error) {
	return group.parent.service.GetEdgeHostnames(contract, group, options)
}

// NewProperty creates a property associated with a given group/contract
func (group *PapiGroup) NewProperty(contract *PapiContract) (*PapiProperty, error) {
	return group.parent.service.NewProperty(contract, group)
}
