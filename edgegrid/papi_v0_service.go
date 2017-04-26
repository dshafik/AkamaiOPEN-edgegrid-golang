package edgegrid

// PapiV0Service is a client for the Property Manager API v0
type PapiV0Service struct {
	client *Client
	config *Config
}

// NewPapiV0Service creates a new PapiV0Service
func NewPapiV0Service(client *Client, config *Config) *PapiV0Service {
	return &PapiV0Service{client: client, config: config}
}

// GetGroups retrieves all groups
func (papi *PapiV0Service) GetGroups() (*PapiGroups, error) {
	groups := NewPapiGroups(papi)
	if err := groups.GetGroups(); err != nil {
		return nil, err
	}

	return groups, nil
}

// GetContracts retrieves all contracts
func (papi *PapiV0Service) GetContracts() (*PapiContracts, error) {
	contracts := NewPapiContracts(papi)
	if err := contracts.GetContracts(); err != nil {
		return nil, err
	}

	return contracts, nil
}

// GetProducts retrieves all products
func (papi *PapiV0Service) GetProducts(contract *PapiContract) (*PapiProducts, error) {
	products := NewPapiProducts(papi)
	if err := products.GetProducts(contract); err != nil {
		return nil, err
	}

	return products, nil
}

// GetEdgeHostnames retrieves all edge hostnames
func (papi *PapiV0Service) GetEdgeHostnames(contract *PapiContract, group *PapiGroup, options string) (*PapiEdgeHostnames, error) {
	edgeHostnames := NewPapiEdgeHostnames(papi)
	if err := edgeHostnames.GetEdgeHostnames(contract, group, options); err != nil {
		return nil, err
	}

	return edgeHostnames, nil
}

// GetCpCodes creates a new PapiCpCodes struct and populates it with all CP Codes associated with a contract/group
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listcpcodes
func (papi *PapiV0Service) GetCpCodes(contract *PapiContract, group *PapiGroup) (*PapiCpCodes, error) {
	cpcodes := NewPapiCpCodes(papi, contract, group)
	if err := cpcodes.GetCpCodes(); err != nil {
		return nil, err
	}

	return cpcodes, nil
}

// GetProperties retrieves all properties for a given contract/group
func (papi *PapiV0Service) GetProperties(contract *PapiContract, group *PapiGroup) (*PapiProperties, error) {
	properties := NewPapiProperties(papi)
	if err := properties.GetProperties(contract, group); err != nil {
		return nil, err
	}

	return properties, nil
}

// GetVersions retrieves all versions for a given property
func (papi *PapiV0Service) GetVersions(property *PapiProperty) (*PapiVersions, error) {
	versions := NewPapiVersions(papi)
	if err := versions.GetVersions(property); err != nil {
		return nil, err
	}

	return versions, nil
}

// GetAvailableBehaviors retrieves all available behaviors for a property
func (papi *PapiV0Service) GetAvailableBehaviors(property *PapiProperty) (*PapiAvailableBehaviors, error) {
	availableBehaviors := NewPapiAvailableBehaviors(papi)
	if err := availableBehaviors.GetAvailableBehaviors(property); err != nil {
		return nil, err
	}

	return availableBehaviors, nil
}

// GetAvailableCriteria retrieves all available criteria for a property
func (papi *PapiV0Service) GetAvailableCriteria(property *PapiProperty) (*PapiAvailableCriteria, error) {
	availableCriteria := NewPapiAvailableCriteria(papi)
	if err := availableCriteria.GetAvailableCriteria(property); err != nil {
		return nil, err
	}

	return availableCriteria, nil
}

// NewProperty creates a new property associated with a given contract/group
func (papi *PapiV0Service) NewProperty(contract *PapiContract, group *PapiGroup) (*PapiProperty, error) {
	if contract == nil {
		contract = NewPapiContract(NewPapiContracts(papi))
		contract.ContractID = group.ContractIDs[0]
	}

	properties := NewPapiProperties(papi)
	property := properties.NewProperty(contract, group)

	return property, nil
}
