package edgegrid

import (
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
	"time"
)

type PapiCpCodes struct {
	service    *PapiV0Service
	AccountId  string `json:"accountId"`
	ContractId string `json:"contractId"`
	GroupId    string `json:"groupId"`
	Cpcodes    struct {
		Items []*PapiCpCode `json:"items"`
	} `json:"cpcodes"`
	Complete chan bool `json:"-"`
}

func (cpcodes *PapiCpCodes) PostUnmarshalJSON() error {
	cpcodes.Init()

	for key, cpcode := range cpcodes.Cpcodes.Items {
		cpcodes.Cpcodes.Items[key].parent = cpcodes

		if cpcode, ok := json.ImplementsPostJsonUnmarshaler(cpcode); ok {
			if err := cpcode.(json.PostJsonUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}

	cpcodes.Complete <- true
	return nil
}

func (cpcodes *PapiCpCodes) NewCpCode() (*PapiCpCode, error) {
	cpcode := &PapiCpCode{parent: cpcodes}
	return cpcode, nil
}

func (cpcodes *PapiCpCodes) Init() {
	cpcodes.Complete = make(chan bool, 1)
}

type PapiCpCode struct {
	parent      *PapiCpCodes
	CpcodeId    string    `json:"cpcodeId,omitempty"`
	CpcodeName  string    `json:"cpcodeName"`
	ProductId   string    `json:"productId,omitempty"`
	ProductIds  []string  `json:"productIds,omitempty"`
	CreatedDate time.Time `json:"createdDate,omitempty"`
	Complete    chan bool `json:"-"`
}

func NewPapiCpCode(parent *PapiCpCodes) *PapiCpCode {
	cpcode := &PapiCpCode{parent: parent}
	cpcode.Init()
	return cpcode
}

func (cpcode *PapiCpCode) Init() {
	cpcode.Complete = make(chan bool, 1)
}

func (cpcode *PapiCpCode) PostUnmarshalJSON() error {
	cpcode.Init()
	cpcode.Complete <- true

	return nil
}
