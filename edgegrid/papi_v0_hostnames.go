package edgegrid

import "github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"

type PapiHostnames struct {
	Resource
	service         *PapiV0Service
	AccountId       string `json:"accountId"`
	ContractId      string `json:"contractId"`
	GroupId         string `json:"groupId"`
	PropertyId      string `json:"propertyId"`
	PropertyVersion int    `json:"propertyVersion"`
	Etag            string `json:"etag"`
	Hostnames       struct {
		Items []*PapiHostname `json:"items"`
	} `json:"hostnames"`
}

func NewPapiHostnames(service *PapiV0Service) *PapiHostnames {
	hostnames := &PapiHostnames{service: service}
	hostnames.Init()

	return hostnames
}

func (hostnames *PapiHostnames) PostUnmarshalJSON() error {
	hostnames.Init()

	for key, hostname := range hostnames.Hostnames.Items {
		hostnames.Hostnames.Items[key].parent = hostnames
		if hostname, ok := json.ImplementsPostJsonUnmarshaler(hostname); ok {
			if err := hostname.(json.PostJsonUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}

	hostnames.Complete <- true

	return nil
}

type PapiHostname struct {
	Resource
	parent         *PapiHostnames
	CnameType      string `json:"cnameType"`
	EdgeHostnameId string `json:"edgeHostnameId"`
	CnameFrom      string `json:"cnameFrom"`
	CnameTo        string `json:"cnameTo,omitempty"`
}
