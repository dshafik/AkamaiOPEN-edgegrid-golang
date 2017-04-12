package edgegrid

import (
	"encoding/json"
	"fmt"
)

type PapiGroups struct {
	service     *PapiV0Service
	AccountId   string `json:"accountId"`
	AccountName string `json:"accountName"`
	Groups      struct {
		Items []*PapiGroup `json:"items"`
	} `json:"groups"`
}

func (groups *PapiGroups) UnmarshalJSON(b []byte) error {
	type PapiGroupsTemp PapiGroups
	temp := &PapiGroupsTemp{service: groups.service}

	if err := json.Unmarshal(b, temp); err != nil {
		return err
	}
	*groups = (PapiGroups)(*temp)

	for key, _ := range groups.Groups.Items {
		groups.Groups.Items[key].parent = groups
	}

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

type PapiGroup struct {
	parent        *PapiGroups
	GroupName     string   `json:"groupName"`
	GroupId       string   `json:"groupId"`
	ParentGroupId string   `json:"parentGroupId,omitempty"`
	ContractIds   []string `json:"contractIds"`
}

func (group *PapiGroup) GetProperties(contract *PapiContract) (*PapiProperties, error) {
	if contract == nil {
		contract = &PapiContract{ContractId: group.ContractIds[0]}
	}
	res, err := group.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties?groupId=%s&contractId=%s",
			group.GroupId,
			contract.ContractId,
		),
	)
	if err != nil {
		return nil, err
	}

	if res.IsError() == true {
		return nil, NewApiError(res)
	}

	properties := &PapiProperties{service: group.parent.service}
	err = res.BodyJson(&properties)
	if err != nil {
		return nil, err
	}

	return properties, nil
}

func (group *PapiGroup) GetCPCodes(contract *PapiContract) (*PapiCpCodes, error) {
	if contract == nil {
		contract = &PapiContract{ContractId: group.ContractIds[0]}
	}
	res, err := group.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/cpcodes?groupId=%s&contractId=%s",
			group.GroupId,
			contract.ContractId,
		),
	)
	if err != nil {
		return nil, err
	}

	if res.IsError() == true {
		return nil, NewApiError(res)
	}

	cpcodes := &PapiCpCodes{service: group.parent.service}
	err = res.BodyJson(&cpcodes)
	if err != nil {
		return nil, err
	}

	return cpcodes, nil
}

func (group *PapiGroup) GetEdgeHostnames(contract *PapiContract, options string) (*PapiEdgeHostnames, error) {
	if contract == nil {
		contract = &PapiContract{ContractId: group.ContractIds[0]}
	}

	if options != "" {
		options = fmt.Sprintf("&options=%s", options)
	}

	res, err := group.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/edgehostnames?groupId=%s&contractId=%s%s",
			group.GroupId,
			contract.ContractId,
			options,
		),
	)
	if err != nil {
		return nil, err
	}

	if res.IsError() == true {
		return nil, NewApiError(res)
	}

	edgeHostnames := &PapiEdgeHostnames{service: group.parent.service}
	err = res.BodyJson(&edgeHostnames)
	if err != nil {
		return nil, err
	}

	return edgeHostnames, nil
}

func (group *PapiGroup) NewProperty(contract *PapiContract) (*PapiProperty, error) {
	if contract == nil {
		contract = &PapiContract{ContractId: group.ContractIds[0]}
	}

	properties := &PapiProperties{service: group.parent.service}

	property := &PapiProperty{
		parent:   properties,
		Contract: contract,
		Group:    group,
	}

	return property, nil
}
