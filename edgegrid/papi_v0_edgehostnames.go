package edgegrid

import "github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"

type PapiEdgeHostnames struct {
	service       *PapiV0Service
	AccountId     string `json:"accountId"`
	ContractId    string `json:"contractId"`
	GroupId       string `json:"groupId"`
	EdgeHostnames struct {
		Items []*PapiEdgeHostname `json:"items"`
	} `json:"edgeHostnames"`
	Complete chan bool `json:"-"`
}

func NewPapiEdgeHostnames(service *PapiV0Service) *PapiEdgeHostnames {
	edgeHostnames := &PapiEdgeHostnames{service: service}
	edgeHostnames.Init()
	return edgeHostnames
}

func (edgeHostnames *PapiEdgeHostnames) Init() {
	edgeHostnames.Complete = make(chan bool, 1)
}

func (edgeHostnames *PapiEdgeHostnames) PostUnmarshalJSON() error {
	edgeHostnames.Init()

	for key, edgeHostname := range edgeHostnames.EdgeHostnames.Items {
		edgeHostnames.EdgeHostnames.Items[key].parent = edgeHostnames

		if edgeHostname, ok := json.ImplementsPostJsonUnmarshaler(edgeHostname); ok {
			if err := edgeHostname.(json.PostJsonUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}

	edgeHostnames.Complete <- true

	return nil
}

type PapiEdgeHostname struct {
	parent             *PapiEdgeHostnames
	EdgeHostnameId     string    `json:"edgeHostnameId,omitempty"`
	EdgeHostnameDomain string    `json:"edgeHostnameDomain,omitempty"`
	ProductId          string    `json:"productId"`
	DomainPrefix       string    `json:"domainPrefix"`
	DomainSuffix       string    `json:"domainSuffix"`
	Status             string    `json:"status,omitempty"`
	Secure             bool      `json:"secure,omitempty"`
	IPVersionBehavior  string    `json:"ipVersionBehavior,omitempty"`
	Complete           chan bool `json:"-"`
}

func NewEdgeHostname(edgeHostnames *PapiEdgeHostnames) (*PapiEdgeHostname, error) {
	edgeHostname := &PapiEdgeHostname{parent: edgeHostnames}
	edgeHostname.Init()
	return edgeHostname, nil
}

func (edgeHostname *PapiEdgeHostname) Init() {
	edgeHostname.Complete = make(chan bool, 1)
}

func (edgeHostname *PapiEdgeHostname) PostUnmarshalJSON() error {
	edgeHostname.Init()
	edgeHostname.Complete <- true
	return nil
}
