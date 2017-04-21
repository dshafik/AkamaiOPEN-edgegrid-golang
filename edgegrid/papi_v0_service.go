package edgegrid

type PapiV0Service struct {
	client *Client
	config *Config
}

func NewPapiV0Service(client *Client, config *Config) *PapiV0Service {
	return &PapiV0Service{client: client, config: config}
}

func (papi *PapiV0Service) GetGroups() (*PapiGroups, error) {
	groups := NewPapiGroups(papi)
	if err := groups.GetGroups(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (papi *PapiV0Service) GetContracts() (*PapiContracts, error) {
	contracts := NewPapiContracts(papi)
	if err := contracts.GetContracts(); err != nil {
		return nil, err
	}

	return contracts, nil
}

func (papi *PapiV0Service) GetProducts(contract *PapiContract) (*PapiProducts, error) {
	products := NewPapiProducts(papi)
	if err := products.GetProducts(contract); err != nil {
		return nil, err
	}

	return products, nil
}

func (papi *PapiV0Service) GetEdgeHostnames(contract *PapiContract, group *PapiGroup, options string) (*PapiEdgeHostnames, error) {
	edgeHostnames := NewPapiEdgeHostnames(papi)
	if err := edgeHostnames.GetEdgeHostnames(contract, group, options); err != nil {
		return nil, err
	}

	return edgeHostnames, nil
}

func (papi *PapiV0Service) GetCpCodes(contract *PapiContract, group *PapiGroup) (*PapiCpCodes, error) {
	cpcodes := NewPapiCpCodes(papi)
	if err := cpcodes.GetCpCodes(contract, group); err != nil {
		return nil, err
	}

	return cpcodes, nil
}

func (papi *PapiV0Service) GetProperties(contract *PapiContract, group *PapiGroup) (*PapiProperties, error) {
	properties := NewPapiProperties(papi)
	if err := properties.GetProperties(contract, group); err != nil {
		return nil, err
	}

	return properties, nil
}

func (papi *PapiV0Service) GetVersions(property *PapiProperty, contract *PapiContract, group *PapiGroup) (*PapiVersions, error) {
	versions := NewPapiVersions(papi)
	if err := versions.GetVersions(property, contract, group); err != nil {
		return nil, err
	}

	return versions, nil
}

func (papi *PapiV0Service) GetAvailableBehaviors(property *PapiProperty, contract *PapiContract, group *PapiGroup) (*PapiAvailableBehaviors, error) {
	availableBehaviors := NewPapiAvailableBehaviors(papi)
	if err := availableBehaviors.GetAvailableBehaviors(property, contract, group); err != nil {
		return nil, err
	}

	return availableBehaviors, nil
}

func (papi *PapiV0Service) GetAvailableCriteria(property *PapiProperty, contract *PapiContract, group *PapiGroup) (*PapiAvailableCriteria, error) {
	availableCriteria := NewPapiAvailableCriteria(papi)
	if err := availableCriteria.GetAvailableCriteria(property, contract, group); err != nil {
		return nil, err
	}

	return availableCriteria, nil
}

func (papi *PapiV0Service) NewProperty(contract *PapiContract, group *PapiGroup) (*PapiProperty, error) {
	if contract == nil {
		contract = NewPapiContract(NewPapiContracts(papi))
		contract.ContractID = group.ContractIDs[0]
	}

	properties := NewPapiProperties(papi)
	property := properties.NewProperty(contract, group)

	return property, nil
}
