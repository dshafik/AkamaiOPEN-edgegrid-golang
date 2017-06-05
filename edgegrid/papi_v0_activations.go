package edgegrid

import (
	gojson "encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"
)

// PapiActivations is a collection of property activations
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

// NewPapiActivations creates a new PapiActivations
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

	if err = res.BodyJSON(activations); err != nil {
		return err
	}

	return nil
}

// GetLatestProductionActivation retrieves the latest activation for the production network
//
// Pass in a status to check for, defaults to PapiStatusActive
func (activations *PapiActivations) GetLatestProductionActivation(status PapiStatusValue) (*PapiActivation, error) {
	return activations.GetLatestActivation(PapiNetworkProduction, status)
}

// GetLatestStagingActivation retrieves the latest activation for the staging network
//
// Pass in a status to check for, defaults to PapiStatusActive
func (activations *PapiActivations) GetLatestStagingActivation(status PapiStatusValue) (*PapiActivation, error) {
	return activations.GetLatestActivation(PapiNetworkStaging, status)
}

// GetLatestActivation gets the latest activation for the specified network
//
// Default to PapiNetworkProduction. Pass in a status to check for, defaults to PapiStatusActive
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

// PapiActivation represents a property activation resource
type PapiActivation struct {
	resource
	parent              *PapiActivations
	ActivationID        string                          `json:"activationId,omitempty"`
	ActivationType      PapiActivationValue             `json:"activationType,omitempty"`
	AcknowledgeWarnings []string                        `json:"acknowledgeWarnings,omitempty"`
	ComplianceRecord    *PapiActivationComplianceRecord `json:"complianceRecord,omitempty"`
	FastPush            bool                            `json:"fastPush,omitempty"`
	IgnoreHTTPErrors    bool                            `json:"ignoreHttpErrors,omitempty"`
	PropertyName        string                          `json:"propertyName,omitempty"`
	PropertyID          string                          `json:"propertyId,omitempty"`
	PropertyVersion     int                             `json:"propertyVersion"`
	Network             PapiNetworkValue                `json:"network"`
	Status              PapiStatusValue                 `json:"status,omitempty"`
	SubmitDate          string                          `json:"submitDate,omitempty"`
	UpdateDate          string                          `json:"updateDate,omitempty"`
	Note                string                          `json:"note,omitempty"`
	NotifyEmails        []string                        `json:"notifyEmails"`
	StatusChange        chan bool                       `json:"-"`
}

type PapiActivationComplianceRecord struct {
	NoncomplianceReason string `json:"noncomplianceReason,omitempty"`
}

// NewPapiActivation creates a new PapiActivation
func NewPapiActivation(parent *PapiActivations) *PapiActivation {
	activation := &PapiActivation{parent: parent}
	activation.Init()

	return activation
}

func (activation *PapiActivation) Init() {
	activation.Complete = make(chan bool, 1)
	activation.StatusChange = make(chan bool, 1)
}

// GetActivation populates the PapiActivation resource
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getanactivation
// Endpoint: GET /papi/v0/properties/{propertyId}/activations/{activationId}{?contractId,groupId}
func (activation *PapiActivation) GetActivation(property *PapiProperty) (time.Duration, error) {
	res, err := activation.parent.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/activations/%s?contractId=%s&groupId=%s",
			property.PropertyID,
			activation.ActivationID,
			property.Contract.ContractID,
			property.Group.GroupID,
		),
	)

	if err != nil {
		return 0, err
	}

	if res.IsError() {
		return 0, NewAPIError(res)
	}

	activations := NewPapiActivations(activation.parent.service)
	if err := res.BodyJSON(activations); err != nil {
		return 0, err
	}

	activation.ActivationID = activations.Activations.Items[0].ActivationID
	activation.ActivationType = activations.Activations.Items[0].ActivationType
	activation.AcknowledgeWarnings = activations.Activations.Items[0].AcknowledgeWarnings
	activation.ComplianceRecord = activations.Activations.Items[0].ComplianceRecord
	activation.FastPush = activations.Activations.Items[0].FastPush
	activation.IgnoreHTTPErrors = activations.Activations.Items[0].IgnoreHTTPErrors
	activation.PropertyName = activations.Activations.Items[0].PropertyName
	activation.PropertyID = activations.Activations.Items[0].PropertyID
	activation.PropertyVersion = activations.Activations.Items[0].PropertyVersion
	activation.Network = activations.Activations.Items[0].Network
	activation.Status = activations.Activations.Items[0].Status
	activation.SubmitDate = activations.Activations.Items[0].SubmitDate
	activation.UpdateDate = activations.Activations.Items[0].UpdateDate
	activation.Note = activations.Activations.Items[0].Note
	activation.NotifyEmails = activations.Activations.Items[0].NotifyEmails

	retry, _ := strconv.Atoi(res.Header.Get("Retry-After"))
	retry *= int(time.Second)

	return time.Duration(retry), nil
}

// Save activates a given property
//
// If acknowledgeWarnings is true and warnings are returned on the first attempt,
// a second attempt is made, acknowledging the warnings.
//
// See: PapiProperty.Activate()
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#activateaproperty
// Endpoint: POST /papi/v0/properties/{propertyId}/activations/{?contractId,groupId}
func (activation *PapiActivation) Save(property *PapiProperty, acknowledgeWarnings bool) error {
	if activation.ComplianceRecord == nil {
		activation.ComplianceRecord = &PapiActivationComplianceRecord{
			NoncomplianceReason: "NO_PRODUCTION_TRAFFIC",
		}
	}

	res, err := activation.parent.service.client.PostJSON(
		fmt.Sprintf(
			"/papi/v0/properties/%s/activations?contractId=%s&groupId=%s",
			property.PropertyID,
			property.Contract.ContractID,
			property.Group.GroupID,
		),
		activation,
	)

	if err != nil {
		return err
	}

	if res.IsError() && (!acknowledgeWarnings || (acknowledgeWarnings && res.StatusCode != 400)) {
		return NewAPIError(res)
	}

	if res.StatusCode == 400 && acknowledgeWarnings {
		warnings := &struct {
			Warnings []struct {
				Detail    string `json:"detail"`
				MessageID string `json:"messageId"`
			} `json:"warnings,omitempty"`
		}{}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		if err = gojson.Unmarshal(body, warnings); err != nil {
			return err
		}

		// Just in case we got a 400 for a different reason
		if len(warnings.Warnings) == 0 {
			jsonBody := &JSONBody{}

			if err = gojson.Unmarshal(body, jsonBody); err != nil {
				return err
			}

			return NewAPIErrorFromBody(res, body)
		}

		for _, warning := range warnings.Warnings {
			activation.AcknowledgeWarnings = append(activation.AcknowledgeWarnings, warning.MessageID)
		}

		// Don't acknowledgeWarnings again, halting a potential endless recursion
		return activation.Save(property, false)
	}

	var location JSONBody
	if err = res.BodyJSON(&location); err != nil {
		return err
	}

	res, err = activation.parent.service.client.Get(
		location["activationLink"].(string),
	)

	activations := NewPapiActivations(activation.parent.service)
	if err := res.BodyJSON(activations); err != nil {
		return err
	}

	activation.ActivationID = activations.Activations.Items[0].ActivationID
	activation.ActivationType = activations.Activations.Items[0].ActivationType
	activation.AcknowledgeWarnings = activations.Activations.Items[0].AcknowledgeWarnings
	activation.ComplianceRecord = activations.Activations.Items[0].ComplianceRecord
	activation.FastPush = activations.Activations.Items[0].FastPush
	activation.IgnoreHTTPErrors = activations.Activations.Items[0].IgnoreHTTPErrors
	activation.PropertyName = activations.Activations.Items[0].PropertyName
	activation.PropertyID = activations.Activations.Items[0].PropertyID
	activation.PropertyVersion = activations.Activations.Items[0].PropertyVersion
	activation.Network = activations.Activations.Items[0].Network
	activation.Status = activations.Activations.Items[0].Status
	activation.SubmitDate = activations.Activations.Items[0].SubmitDate
	activation.UpdateDate = activations.Activations.Items[0].UpdateDate
	activation.Note = activations.Activations.Items[0].Note
	activation.NotifyEmails = activations.Activations.Items[0].NotifyEmails

	return nil
}

// PollStatus will responsibly poll till the property is active or an error occurs
//
// The PapiActivation.StatusChange is a channel that can be used to
// block on status changes. If a new valid status is returned, true will
// be sent to the channel, otherwise, false will be sent.
//
//	go activation.PollStatus(property)
//	for activation.Status != edgegrid.PapiStatusActive {
//		select {
//		case statusChanged := <-activation.StatusChange:
//			if statusChanged == false {
//				break
//			}
//		case <-time.After(time.Minute * 30):
//			break
//		}
//	}
//
//	if activation.Status == edgegrid.PapiStatusActive {
//		// Activation succeeded
//	}
func (activation *PapiActivation) PollStatus(property *PapiProperty) bool {
	currentStatus := activation.Status
	var retry time.Duration = 0

	for currentStatus != PapiStatusActive {
		time.Sleep(retry)

		var err error
		retry, err = activation.GetActivation(property)

		if err != nil {
			activation.StatusChange <- false
			return false
		}

		if activation.Network == PapiNetworkStaging && retry > time.Minute {
			retry = time.Minute
		}

		if err != nil {
			activation.StatusChange <- false
			return false
		}

		if currentStatus != activation.Status {
			currentStatus = activation.Status
			activation.StatusChange <- true
		}
	}

	return true
}

// Cancel an activation in progress
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#cancelapendingactivation
// Endpoint: DELETE /papi/v0/properties/{propertyId}/activations/{activationId}{?contractId,groupId}
func (activation *PapiActivation) Cancel(property *PapiProperty) error {
	res, err := activation.parent.service.client.Delete(
		fmt.Sprintf(
			"/papi/v0/properties/%s/activations?contractId=%s&groupId=%s",
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

	newActivations := NewPapiActivations(activation.parent.service)
	if err := res.BodyJSON(newActivations); err != nil {
		return err
	}

	activation.ActivationID = newActivations.Activations.Items[0].ActivationID
	activation.ActivationType = newActivations.Activations.Items[0].ActivationType
	activation.AcknowledgeWarnings = newActivations.Activations.Items[0].AcknowledgeWarnings
	activation.ComplianceRecord = newActivations.Activations.Items[0].ComplianceRecord
	activation.FastPush = newActivations.Activations.Items[0].FastPush
	activation.IgnoreHTTPErrors = newActivations.Activations.Items[0].IgnoreHTTPErrors
	activation.PropertyName = newActivations.Activations.Items[0].PropertyName
	activation.PropertyID = newActivations.Activations.Items[0].PropertyID
	activation.PropertyVersion = newActivations.Activations.Items[0].PropertyVersion
	activation.Network = newActivations.Activations.Items[0].Network
	activation.Status = newActivations.Activations.Items[0].Status
	activation.SubmitDate = newActivations.Activations.Items[0].SubmitDate
	activation.UpdateDate = newActivations.Activations.Items[0].UpdateDate
	activation.Note = newActivations.Activations.Items[0].Note
	activation.NotifyEmails = newActivations.Activations.Items[0].NotifyEmails
	activation.StatusChange = newActivations.Activations.Items[0].StatusChange

	return nil
}

// PapiActivationValue is used to create an "enum" of possible PapiActivation.ActivationType values
type PapiActivationValue string

// PapiNetworkValue is used to create an "enum" of possible PapiActivation.Network values
type PapiNetworkValue string

// PapiStatusValue is used to create an "enum" of possible PapiActivation.Status values
type PapiStatusValue string

const (
	// PapiActivationTypeActivate PapiActivation.ActivationType value ACTIVATE
	PapiActivationTypeActivate PapiActivationValue = "ACTIVATE"
	// PapiActivationTypeDeactivate PapiActivation.ActivationType value DEACTIVATE
	PapiActivationTypeDeactivate PapiActivationValue = "DEACTIVATE"

	// PapiNetworkProduction PapiActivation.Network value PRODUCTION
	PapiNetworkProduction PapiNetworkValue = "PRODUCTION"
	// PapiNetworkStaging PapiActivation.Network value STAGING
	PapiNetworkStaging PapiNetworkValue = "STAGING"

	// PapiStatusActive PapiActivation.Status value ACTIVE
	PapiStatusActive PapiStatusValue = "ACTIVE"
	// PapiStatusInactive PapiActivation.Status value INACTIVE
	PapiStatusInactive PapiStatusValue = "INACTIVE"
	// PapiStatusPending PapiActivation.Status value PENDING
	PapiStatusPending PapiStatusValue = "PENDING"
	// PapiStatusZone1 PapiActivation.Status value ZONE_1
	PapiStatusZone1 PapiStatusValue = "ZONE_1"
	// PapiStatusZone2 PapiActivation.Status value ZONE_2
	PapiStatusZone2 PapiStatusValue = "ZONE_2"
	// PapiStatusZone3 PapiActivation.Status value ZONE_3
	PapiStatusZone3 PapiStatusValue = "ZONE_3"
	// PapiStatusAborted PapiActivation.Status value ABORTED
	PapiStatusAborted PapiStatusValue = "ABORTED"
	// PapiStatusFailed PapiActivation.Status value FAILED
	PapiStatusFailed PapiStatusValue = "FAILED"
	// PapiStatusDeactivated PapiActivation.Status value DEACTIVATED
	PapiStatusDeactivated PapiStatusValue = "DEACTIVATED"
	// PapiStatusPendingDeactivation PapiActivation.Status value PENDING_DEACTIVATION
	PapiStatusPendingDeactivation PapiStatusValue = "PENDING_DEACTIVATION"
	// PapiStatusNew PapiActivation.Status value NEW
	PapiStatusNew PapiStatusValue = "NEW"
)
