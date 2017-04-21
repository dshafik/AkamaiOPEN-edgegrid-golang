package edgegrid

import (
	"fmt"
	"time"
)

type PapiActivations struct {
	Resource
	service     *PapiV0Service
	AccountId   string `json:"accountId"`
	ContractId  string `json:"contractId"`
	GroupId     string `json:"groupId"`
	Activations struct {
		Items []*PapiActivation `json:"items"`
	} `json:"activations"`
}

func NewPapiActivations(service *PapiV0Service) *PapiActivations {
	activations := &PapiActivations{service: service}
	activations.Init()

	return activations
}

func (activations *PapiActivations) GetLatestProductionActivation(status PapiStatusValue) (*PapiActivation, error) {
	return activations.GetLatestActivation(PAPI_NETWORK_PRODUCTION, status)
}

func (activations *PapiActivations) GetLatestStagingActivation(status PapiStatusValue) (*PapiActivation, error) {
	return activations.GetLatestActivation(PAPI_NETWORK_STAGING, status)
}

func (activations *PapiActivations) GetLatestActivation(network PapiNetworkValue, status PapiStatusValue) (*PapiActivation, error) {
	if network == "" {
		network = PAPI_NETWORK_PRODUCTION
	}

	if status == "" {
		status = PAPI_STATUS_ACTIVE
	}

	var latest *PapiActivation
	for _, activation := range activations.Activations.Items {
		if activation.Network == network && activation.Status == status && (latest == nil || activation.PropertyVersion > latest.PropertyVersion) {
			latest = activation
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("No activation found (network: %s, status: %s)", network, status)
	}

	return latest, nil
}

type PapiActivation struct {
	Resource
	parent              *PapiActivations
	ActivationId        string              `json:"activationId,omitempty"`
	ActivationType      PapiActivationValue `json:"activationType,omitempty"`
	AcknowledgeWarnings []string            `json:"acknowledgeWarnings,omitempty"`
	FastPush            bool                `json:"fastPush,omitempty"`
	IgnoreHttpErrors    bool                `json:"ignoreHttpErrors,omitempty"`
	PropertyName        string              `json:"propertyName,omitempty"`
	PropertyId          string              `json:"propertyId,omitempty"`
	PropertyVersion     int                 `json:"propertyVersion"`
	Network             PapiNetworkValue    `json:"network"`
	Status              PapiStatusValue     `json:"status,omitempty"`
	SubmitDate          time.Time           `json:"submitDate,omitempty"`
	UpdateDate          time.Time           `json:"updateDate,omitempty"`
	Note                string              `json:"note,omitempty"`
	NotifyEmails        []string            `json:"notifyEmails"`
}

func NewPapiActivation(parent *PapiActivations) *PapiActivation {
	activation := &PapiActivation{parent: parent}
	activation.Init()

	return activation
}

type PapiActivationValue string
type PapiNetworkValue string
type PapiStatusValue string

const (
	PAPI_ACTIVATION_TYPE_ACTIVATE    PapiActivationValue = "ACTIVATE"
	PAPI_ACTIVATION_TYPE_DEACTIVATE  PapiActivationValue = "DEACTIVATE"
	PAPI_NETWORK_PRODUCTION          PapiNetworkValue    = "PRODUCTION"
	PAPI_NETWORK_STAGING             PapiNetworkValue    = "STAGING"
	PAPI_STATUS_ACTIVE               PapiStatusValue     = "ACTIVE"
	PAPI_STATUS_INACTIVE             PapiStatusValue     = "INACTIVE"
	PAPI_STATUS_PENDING              PapiStatusValue     = "PENDING"
	PAPI_STATUS_ZONE_1               PapiStatusValue     = "ZONE_1"
	PAPI_STATUS_ZONE_2               PapiStatusValue     = "ZONE_2"
	PAPI_STATUS_ZONE_3               PapiStatusValue     = "ZONE_3"
	PAPI_STATUS_ABORTED              PapiStatusValue     = "ABORTED"
	PAPI_STATUS_FAILED               PapiStatusValue     = "FAILED"
	PAPI_STATUS_DEACTIVATED          PapiStatusValue     = "DEACTIVATED"
	PAPI_STATUS_PENDING_DEACTIVATION PapiStatusValue     = "PENDING_DEACTIVATION"
	PAPI_STATUS_NEW                  PapiStatusValue     = "NEW"
)
