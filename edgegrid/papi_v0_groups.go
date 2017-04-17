package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

type PapiGroups struct {
	service     *PapiV0Service
	AccountId   string `json:"accountId"`
	AccountName string `json:"accountName"`
	Groups      struct {
		Items []*PapiGroup `json:"items"`
	} `json:"groups"`
	Complete chan bool `json:"-"`
}

func (groups *PapiGroups) Init() {
	groups.Complete = make(chan bool, 1)
}

func (groups *PapiGroups) PostUnmarshalJSON() error {
	groups.Init()

	for key, group := range groups.Groups.Items {
		groups.Groups.Items[key].parent = groups
		if group, ok := json.ImplementsPostJsonUnmarshaler(group); ok {
			if err := group.(json.PostJsonUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}

	groups.Complete <- true

	return nil
}

func (groups *PapiGroups) AddGroup(newGroup *PapiGroup) {
	if newGroup.GroupId != "" {
		for key, group := range groups.Groups.Items {
			if group.GroupId == newGroup.GroupId {
				groups.Groups.Items[key] = newGroup
				return
			}
		}
	}

	newGroup.parent = groups

	groups.Groups.Items = append(groups.Groups.Items, newGroup)
}

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

type PapiGroup struct {
	parent        *PapiGroups
	GroupName     string    `json:"groupName"`
	GroupId       string    `json:"groupId"`
	ParentGroupId string    `json:"parentGroupId,omitempty"`
	ContractIds   []string  `json:"contractIds"`
	Complete      chan bool `json:"-"`
}

func NewPapiGroup(parent *PapiGroups) *PapiGroup {
	group := &PapiGroup{
		parent: parent,
	}
	group.Init()
	return group
}

func (group *PapiGroup) Init() {
	group.Complete = make(chan bool, 1)
}

func (group *PapiGroup) PostUnmarshalJSON() error {
	group.Init()
	group.Complete <- true
	return nil
}

func (group *PapiGroup) GetProperties(contract *PapiContract) (*PapiProperties, error) {
	return group.parent.service.GetProperties(contract, group)
}

func (group *PapiGroup) GetCPCodes(contract *PapiContract) (*PapiCpCodes, error) {
	return group.parent.service.GetCPCodes(contract, group)
}

func (group *PapiGroup) GetEdgeHostnames(contract *PapiContract, options string) (*PapiEdgeHostnames, error) {
	return group.parent.service.GetEdgeHostnames(contract, group, options)
}

func (group *PapiGroup) NewProperty(contract *PapiContract) (*PapiProperty, error) {
	return group.parent.service.NewProperty(contract, group)
}
