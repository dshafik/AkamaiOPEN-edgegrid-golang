package edgegrid

import (
	"fmt"
	"time"
)

type PapiActivations struct {
	resource
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

// GetActivations retrieves activation data for a given property
//
// See: PapiProperty.GetActivations()
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listactivations
// Endpoint: GET /papi/v0/properties/{propertyId}/activations/{?contractId,groupId}
func (activations *PapiActivations) GetActivations(property *PapiProperty) error {
	res, err := activations.service.client.Get(
		fmt.Sprintf("/papi/v0/properties/%s/activations?contractId=%s&groupId=%s",
			property.PropertyID,
			property.Contract.ContractID,
			property.Group.GroupID,
		),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newActivations := NewPapiActivations(property.parent.service)
	if err = res.BodyJSON(activations); err != nil {
		return err
	}

	*activations = *newActivations

	return nil
}

func (activations *PapiActivations) GetLatestProductionActivation(status PapiStatusValue) (*PapiActivation, error) {
	return activations.GetLatestActivation(PapiNetworkProduction, status)
}

func (activations *PapiActivations) GetLatestStagingActivation(status PapiStatusValue) (*PapiActivation, error) {
	return activations.GetLatestActivation(PapiNetworkStaging, status)
}

func (activations *PapiActivations) GetLatestActivation(network PapiNetworkValue, status PapiStatusValue) (*PapiActivation, error) {
	if network == "" {
		network = PapiNetworkProduction
	}

	if status == "" {
		status = PapiStatusActive
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
	resource
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
	PapiActivationTypeActivate    PapiActivationValue = "ACTIVATE"
	PapiActivationTypeDeactivate  PapiActivationValue = "DEACTIVATE"
	PapiNetworkProduction         PapiNetworkValue    = "PRODUCTION"
	PapiNetworkStaging            PapiNetworkValue    = "STAGING"
	PapiStatusActive              PapiStatusValue     = "ACTIVE"
	PapiStatusInactive            PapiStatusValue     = "INACTIVE"
	PapiStatusPending             PapiStatusValue     = "PENDING"
	PapiStatusZone1               PapiStatusValue     = "ZONE_1"
	PapiStatusZone2               PapiStatusValue     = "ZONE_2"
	PapiStatusZone3               PapiStatusValue     = "ZONE_3"
	PapiStatusAborted             PapiStatusValue     = "ABORTED"
	PapiStatusFailed              PapiStatusValue     = "FAILED"
	PapiStatusDeactivated         PapiStatusValue     = "DEACTIVATED"
	PapiStatusPendingDeactivation PapiStatusValue     = "PENDING_DEACTIVATION"
	PapiStatusNew                 PapiStatusValue     = "NEW"
)
