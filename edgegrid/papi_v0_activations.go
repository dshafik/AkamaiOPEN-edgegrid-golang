package edgegrid

import (
	"fmt"
	"time"
)

type PapiActivations struct {
	Resource
	service     *PapiV0Service
	AccountID   string `json:"accountId"`
	ContractID  string `json:"contractId"`
	GroupID     string `json:"groupId"`
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
	return activations.GetLatestActivation(papiNetworkProduction, status)
}

func (activations *PapiActivations) GetLatestStagingActivation(status PapiStatusValue) (*PapiActivation, error) {
	return activations.GetLatestActivation(papiNetworkStaging, status)
}

func (activations *PapiActivations) GetLatestActivation(network PapiNetworkValue, status PapiStatusValue) (*PapiActivation, error) {
	if network == "" {
		network = papiNetworkProduction
	}

	if status == "" {
		status = papiStatusActive
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
	ActivationID        string              `json:"activationId,omitempty"`
	ActivationType      PapiActivationValue `json:"activationType,omitempty"`
	AcknowledgeWarnings []string            `json:"acknowledgeWarnings,omitempty"`
	FastPush            bool                `json:"fastPush,omitempty"`
	IgnoreHTTPErrors    bool                `json:"ignoreHttpErrors,omitempty"`
	PropertyName        string              `json:"propertyName,omitempty"`
	PropertyID          string              `json:"propertyId,omitempty"`
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
	papiActivationTypeActivate    PapiActivationValue = "ACTIVATE"
	papiActivationTypeDeactivate  PapiActivationValue = "DEACTIVATE"
	papiNetworkProduction         PapiNetworkValue    = "PRODUCTION"
	papiNetworkStaging            PapiNetworkValue    = "STAGING"
	papiStatusActive              PapiStatusValue     = "ACTIVE"
	papiStatusInactive            PapiStatusValue     = "INACTIVE"
	papiStatusPending             PapiStatusValue     = "PENDING"
	papiStatusZone1               PapiStatusValue     = "ZONE_1"
	papiStatusZone2               PapiStatusValue     = "ZONE_2"
	papiStatusZone3               PapiStatusValue     = "ZONE_3"
	papiStatusAborted             PapiStatusValue     = "ABORTED"
	papiStatusFailed              PapiStatusValue     = "FAILED"
	papiStatusDeactivated         PapiStatusValue     = "DEACTIVATED"
	papiStatusPendingDeactivation PapiStatusValue     = "PENDING_DEACTIVATION"
	papiStatusNew                 PapiStatusValue     = "NEW"
)
