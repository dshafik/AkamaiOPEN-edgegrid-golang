package edgegrid

import (
	"encoding/json"
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
}

func (cpcodes *PapiCpCodes) UnmarshalJSON(b []byte) error {
	type PapiCpCodesTemp PapiCpCodes
	temp := &PapiCpCodesTemp{service: cpcodes.service}

	if err := json.Unmarshal(b, temp); err != nil {
		return err
	}
	*cpcodes = (PapiCpCodes)(*temp)

	for key, _ := range cpcodes.Cpcodes.Items {
		cpcodes.Cpcodes.Items[key].parent = cpcodes
	}

	return nil
}

func (cpcodes *PapiCpCodes) NewCpCode() (*PapiCpCode, error) {
	cpcode := &PapiCpCode{parent: cpcodes}
	return cpcode, nil
}

type PapiCpCode struct {
	parent      *PapiCpCodes
	CpcodeId    string    `json:"cpcodeId,omitempty"`
	CpcodeName  string    `json:"cpcodeName"`
	ProductId   string    `json:"productId,omitempty"`
	ProductIds  []string  `json:"productIds,omitempty"`
	CreatedDate time.Time `json:"createdDate,omitempty"`
}
