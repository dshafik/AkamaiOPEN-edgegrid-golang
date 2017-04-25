package edgegrid

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// PapiRules is a collection of property rules
type PapiRules struct {
	resource
	service         *PapiV0Service
	AccountID       string            `json:"accountId"`
	ContractID      string            `json:"contractId"`
	GroupID         string            `json:"groupId"`
	PropertyID      string            `json:"propertyId"`
	PropertyVersion int               `json:"propertyVersion"`
	Etag            string            `json:"etag"`
	RuleFormat      string            `json:"ruleFormat"`
	Rules           *PapiRule         `json:"rules"`
	Errors          []*PapiRuleErrors `json:"errors,omitempty"`
}

// NewPapiRules creates a new PapiRules
func NewPapiRules(service *PapiV0Service) *PapiRules {
	rules := &PapiRules{service: service}
	rules.Init()

	return rules
}

// PostUnmarshalJSON is called after JSON unmarshaling into PapiEdgeHostnames
//
// See: edgegrid/json.Unmarshal()
func (rules *PapiRules) PostUnmarshalJSON() error {
	rules.Init()

	for key := range rules.Rules.Behaviors {
		rules.Rules.Behaviors[key].parent = rules.Rules
		if len(rules.Rules.Children) > 0 {
			for _, v := range rules.Rules.GetChildren(0, 0) {
				for _, j := range v.Behaviors {
					j.parent = rules.Rules
				}
			}
		}
	}

	for key := range rules.Rules.Criteria {
		rules.Rules.Criteria[key].parent = rules.Rules
	}

	return nil
}

// PreMarshalJSON is called before JSON marshaling
//
// See: edgegrid/json.Marshal()
func (rules *PapiRules) PreMarshalJSON() error {
	rules.Errors = nil
	return nil
}

// PrintRules prints a reasonably easy to read tree of all rules and behaviors on a property
func (rules *PapiRules) PrintRules() error {
	group := NewPapiGroup(NewPapiGroups(rules.service))
	group.GroupID = rules.GroupID
	group.ContractIDs = []string{rules.ContractID}

	properties, _ := group.GetProperties(nil)
	var property *PapiProperty
	for _, property = range properties.Properties.Items {
		if property.PropertyID == rules.PropertyID {
			break
		}
	}

	fmt.Println(property.PropertyName)

	fmt.Println("├── Criteria")
	for _, criteria := range rules.Rules.Criteria {
		fmt.Printf("│   ├── %s\n", criteria.Name)
		i := 0
		for option, value := range *criteria.Options {
			i++
			if i < len(*criteria.Options) {
				fmt.Printf("│   │   ├── %s: %#v\n", option, value)
			} else {
				fmt.Printf("│   │   └── %s: %#v\n", option, value)
			}
		}
	}

	fmt.Println("└── Behaviors")

	prefix := "   │"
	i := 0
	for _, behavior := range rules.Rules.Behaviors {
		i++
		if i < len(rules.Rules.Behaviors) && len(rules.Rules.Children) != 0 {
			fmt.Printf("   ├── Behavior: %s\n", behavior.Name)
		} else {
			fmt.Printf("   └── Behavior: %s\n", behavior.Name)
		}

		j := 0

		for option, value := range *behavior.Options {
			j++
			if i == len(rules.Rules.Behaviors) && len(rules.Rules.Children) == 0 {
				prefix = strings.TrimSuffix(prefix, "│")
			}

			if j < len(*behavior.Options) {
				fmt.Printf("%s   ├── Option: %s: %#v\n", prefix, option, value)
			} else {
				fmt.Printf("%s   └── Option: %s: %#v\n", prefix, option, value)
			}
		}
	}

	if len(rules.Rules.Children) > 0 {
		i := 0
		children := rules.Rules.GetChildren(0, 0)
		for _, child := range children {
			i++
			spacer := strings.TrimSuffix(strings.Repeat(prefix, child.Depth), "│")
			if i < len(children) {
				fmt.Printf("%s├── Section: %s\n", spacer, child.Name)
			} else {
				fmt.Printf("%s└── Section: %s\n", spacer, child.Name)
			}

			spacer = strings.TrimSuffix(strings.Repeat(prefix, child.Depth+1), "│")
			j := 0
			for _, behavior := range child.Behaviors {
				j++
				if j < len(child.Behaviors) {
					fmt.Printf("%s├── Behavior: %s\n", spacer, behavior.Name)
				} else {
					//spacer = strings.TrimSuffix(spacer, "│   ") + "    "
					fmt.Printf("%s└── Behavior: %s\n", spacer, behavior.Name)
				}
				space := strings.TrimSuffix(strings.Repeat(prefix, child.Depth+2), "│")

				fmt.Printf("%s├── Criteria\n", space)
				i := 0
				for _, criteria := range child.Criteria {
					i++
					if i < len(child.Criteria) {
						fmt.Printf("   │%s├── %s\n", space, criteria.Name)
					} else {
						fmt.Printf("   │%s└── %s\n", space, criteria.Name)
					}
					k := 0
					for option, value := range *criteria.Options {
						k++
						if k < len(*criteria.Options) {
							fmt.Printf("   │   │%s├── %s: %#v\n", space, option, value)
						} else {
							fmt.Printf("   │   │%s└── %s: %#v\n", space, option, value)
						}
					}
				}

				k := 0
				for option, value := range *behavior.Options {
					k++
					if k < len(*behavior.Options) {
						fmt.Printf("%s├── Option: %s: %#v\n", space, option, value)
					} else {
						fmt.Printf("%s└── Option: %s: %#v\n", space, option, value)
					}
				}
			}
		}
	}

	return nil
}

// GetRules populates PapiRules with rule data for a given property
//
// See: PapiProperty.GetRules
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getaruletree
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/{propertyVersion}/rules/{?contractId,groupId}
func (rules *PapiRules) GetRules(property *PapiProperty) error {
	res, err := rules.service.client.Get(fmt.Sprintf(
		"/papi/v0/properties/%s/versions/%d/rules?contractId=%s&groupId=%s",
		property.PropertyID,
		property.LatestVersion,
		property.Contract.ContractID,
		property.Group.GroupID,
	))

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	newRules := NewPapiRules(property.parent.service)
	if err = res.BodyJSON(newRules); err != nil {
		return err
	}

	*rules = *newRules

	return nil
}

// GetRulesDigest fetches the Etag for a rule tree
//
// See: PapiProperty.GetRulesDigest()
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#getaruletreesdigest
// Endpoint: HEAD /papi/v0/properties/{propertyId}/versions/{propertyVersion}/rules/{?contractId,groupId}
func (rules *PapiRules) GetRulesDigest(property *PapiProperty) (string, error) {
	res, err := rules.service.client.Head(fmt.Sprintf(
		"/papi/v0/properties/%s/versions/%d/rules?contractId=%s&groupId=%s",
		property.PropertyID,
		property.LatestVersion,
		property.Contract.ContractID,
		property.Group.GroupID,
	))

	if err != nil {
		return "", err
	}

	if res.IsError() {
		return "", NewAPIError(res)
	}

	return res.Header.Get("Etag"), nil
}

// GetAllRules returns a flattened rule tree for easy iteration
//
// Each PapiRule has a Depth property that makes it possible to identify
// it's original placement in the tree.
func (rules *PapiRules) GetAllRules() []*PapiRule {
	var flatRules []*PapiRule
	flatRules = append(flatRules, rules.Rules)
	flatRules = append(flatRules, rules.Rules.GetChildren(0, 0)...)

	return flatRules
}

// Save creates/updates a rule tree for a property
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#updatearuletree
// Endpoint: PUT /papi/v0/properties/{propertyId}/versions/{propertyVersion}/rules/{?contractId,groupId}
func (rules *PapiRules) Save() error {
	// /papi/v0/properties/{propertyId}/versions/{propertyVersion}/rules/{?contractId,groupId}
	res, err := rules.service.client.PutJSON(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions/%d/rules/?contractId=%s&groupId=%s",
			rules.PropertyID,
			rules.PropertyVersion,
			rules.ContractID,
			rules.GroupID,
		),
		rules,
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewAPIError(res)
	}

	if err = res.BodyJSON(rules); err != nil {
		return err
	}

	if len(rules.Errors) != 0 {
		return fmt.Errorf("there were %d errors. See rules.Errors for details", len(rules.Errors))
	}

	return nil
}

// SetBehaviorOptions sets the options on a given behavior path
//
// path is a / delimited path from the root of the rule set to the behavior.
// All existing options are overwritten. To add/replace options see AddBehaviorOptions
// instead.
//
// For example, to set the CP Code, the behavior exists at the root, and is called "cpCode",
// making the path, "/cpCode":
//
//	rules.SetBehaviorOptions(
//		"/cpCode",
//		edgegrid.PapiOptionValue{
//			"value": edgegrid.PapiOptionValue{
//				"id": cpcode,
//			},
//		},
//	)
//
// The adaptiveImageCompression behavior defaults to being under the Performance -> JPEG Images,
// making the path "/Performance/JPEG Images/adaptiveImageCompression":
//
//	rules.SetBehaviorOptions(
//		"/Performance/JPEG Images/adaptiveImageCompression"",
//		edgegrid.PapiOptionValue{
//			"tier3StandardCompressionValue": 30,
//		},
//	)
//
// However, this would replace all other options for that behavior, meaning the example
// would fail to validate without the other required options.
func (rules *PapiRules) SetBehaviorOptions(path string, newOptions PapiOptionValue) error {
	behavior, err := rules.FindBehavior(path)
	if err != nil {
		return err
	}

	behavior.Options = &newOptions
	return nil
}

// AddBehaviorOptions adds/replaces options on a given behavior path
//
// path is a / delimited path from the root of the rule set to the behavior.
// Individual existing options are overwritten. To replace all options, see SetBehaviorOptions
// instead.
//
// For example, to change the CP Code, the behavior exists at the root, and is called "cpCode",
// making the path, "/cpCode":
//
//	rules.AddBehaviorOptions(
//		"/cpCode",
//		edgegrid.PapiOptionValue{
//			"value": edgegrid.PapiOptionValue{
//				"id": cpcode,
//			},
//		},
//	)
//
// The adaptiveImageCompression behavior defaults to being under the Performance -> JPEG Images,
// making the path "/Performance/JPEG Images/adaptiveImageCompression":
//
//	rules.AddBehaviorOptions(
//		"/Performance/JPEG Images/adaptiveImageCompression"",
//		edgegrid.PapiOptionValue{
//			"tier3StandardCompressionValue": 30,
//		},
//	)
//
// This will only change the "tier3StandardCompressionValue" option value.
func (rules *PapiRules) AddBehaviorOptions(path string, newOptions PapiOptionValue) error {
	behavior, err := rules.FindBehavior(path)
	if err != nil {
		return err
	}

	options := *behavior.Options
	for key, value := range newOptions {
		options[key] = value
	}
	behavior.Options = &options

	return nil
}

// FindBehavior locates a specific behavior by path
//
// See SetBehaviorOptions and AddBehaviorOptions for examples of paths.
func (rules *PapiRules) FindBehavior(path string) (*PapiBehavior, error) {
	if len(path) <= 1 {
		return nil, fmt.Errorf("Invalid Path: \"%s\"", path)
	}

	sep := "/"
	segments := strings.Split(strings.ToLower(strings.TrimPrefix(path, sep)), sep)

	if len(segments) == 1 {
		for _, behavior := range rules.Rules.Behaviors {
			if strings.ToLower(behavior.Name) == segments[0] {
				return behavior, nil
			}
		}
		return nil, fmt.Errorf("Path not found: \"%s\"", path)
	}

	currentRule := rules.Rules
	i := 0
	for _, segment := range segments {
		i++
		if i < len(segments) {
			for _, rule := range currentRule.GetChildren(0, 1) {
				if strings.ToLower(rule.Name) == segment {
					currentRule = rule
				}
			}
		} else {
			for _, behavior := range currentRule.Behaviors {
				if strings.ToLower(behavior.Name) == segment {
					return behavior, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("Path not found: \"%s\"", path)
}

// PapiRule represents a property rule resource
type PapiRule struct {
	resource
	parent         *PapiRules
	Depth          int             `json:"-"`
	Name           string          `json:"name"`
	Criteria       []*PapiCriteria `json:"criteria,omitempty"`
	Behaviors      []*PapiBehavior `json:"behaviors,omitempty"`
	Children       []*PapiRule     `json:"children,omitempty"`
	Comment        string          `json:"comment,omitempty"`
	CriteriaLocked bool            `json:"criteriaLocked,omitempty"`
	Options        struct {
		IsSecure bool `json:"is_secure,omitempty"`
	} `json:"options,omitempty"`
}

// NewPapiRule creates a new PapiRule
func NewPapiRule(parent *PapiRules) *PapiRule {
	rule := &PapiRule{parent: parent}
	rule.Init()

	return rule
}

// GetChildren recurses a Rule tree and retrieves all child rules
func (rule *PapiRule) GetChildren(depth int, limit int) []*PapiRule {
	depth++

	if limit != 0 && depth > limit {
		return nil
	}

	var children []*PapiRule
	if len(rule.Children) > 0 {
		for _, v := range rule.Children {
			v.Depth = depth
			children = append(children, v)
			children = append(children, v.GetChildren(depth, limit)...)
		}
	}

	return children
}

// AddChildRule appends a child rule
func (rule *PapiRule) AddChildRule(child *PapiRule) {
	rule.Children = append(rule.Children, child)
}

// AddCriteria appends a rule criteria
func (rule *PapiRule) AddCriteria(critera *PapiCriteria) {
	rule.Criteria = append(rule.Criteria, critera)
}

// AddBehavior appends a rule behavior
func (rule *PapiRule) AddBehavior(behavior *PapiBehavior) {
	rule.Behaviors = append(rule.Behaviors, behavior)
}

// PapiCriteria represents a rule criteria resource
type PapiCriteria struct {
	resource
	parent  *PapiRule
	Name    string           `json:"name"`
	Options *PapiOptionValue `json:"options"`
}

// NewPapiCriteria creates a new PapiCriteria
func NewPapiCriteria(parent *PapiRule) *PapiCriteria {
	criteria := &PapiCriteria{parent: parent}
	criteria.Init()

	return criteria
}

// PapiBehavior represents a rule behavior resource
type PapiBehavior struct {
	resource
	parent  *PapiRule
	Name    string           `json:"name"`
	Options *PapiOptionValue `json:"options"`
}

// NewPapiBehavior creates a new PapiBehavior
func NewPapiBehavior(parent *PapiRule) *PapiBehavior {
	behavior := &PapiBehavior{parent: parent}
	behavior.Init()

	return behavior
}

// PapiOptionValue represents a generic option value
//
// PapiOptionValue is a map with string keys, and any
// type of value. You can nest PapiOptionValues as necessary
// to create more complex values.
type PapiOptionValue JSONBody

// PapiAvailableCriteria represents a collection of available rule criteria
type PapiAvailableCriteria struct {
	resource
	service           *PapiV0Service
	ContractID        string `json:"contractId"`
	GroupID           string `json:"groupId"`
	ProductID         string `json:"productId"`
	RuleFormat        string `json:"ruleFormat"`
	AvailableCriteria struct {
		Items []struct {
			Name       string `json:"name"`
			SchemaLink string `json:"schemaLink"`
		} `json:"items"`
	} `json:"availableCriteria"`
}

// NewPapiAvailableCriteria creates a new PapiAvailableCriteria
func NewPapiAvailableCriteria(service *PapiV0Service) *PapiAvailableCriteria {
	availableCriteria := &PapiAvailableCriteria{service: service}
	availableCriteria.Init()

	return availableCriteria
}

// GetAvailableCriteria retrieves criteria available for a given property
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listavailablecriteria
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/{propertyVersion}/available-criteria{?contractId,groupId}
func (availableCriteria *PapiAvailableCriteria) GetAvailableCriteria(property *PapiProperty) error {
	res, err := availableCriteria.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions/%d/available-criteria?contractId=%s&groupId=%s",
			property.PropertyID,
			property.LatestVersion,
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

	newAvailableCriteria := NewPapiAvailableCriteria(availableCriteria.service)
	if err = res.BodyJSON(newAvailableCriteria); err != nil {
		return err
	}

	*availableCriteria = *newAvailableCriteria

	return nil
}

// PapiAvailableBehaviors represents a collection of available rule behaviors
type PapiAvailableBehaviors struct {
	resource
	service    *PapiV0Service
	ContractID string `json:"contractId"`
	GroupID    string `json:"groupId"`
	ProductID  string `json:"productId"`
	RuleFormat string `json:"ruleFormat"`
	Behaviors  struct {
		Items []PapiAvailableBehavior `json:"items"`
	} `json:"behaviors"`
}

// NewPapiAvailableBehaviors creates a new PapiAvailableBehaviors
func NewPapiAvailableBehaviors(service *PapiV0Service) *PapiAvailableBehaviors {
	availableBehaviors := &PapiAvailableBehaviors{service: service}
	availableBehaviors.Init()

	return availableBehaviors
}

// PostUnmarshalJSON is called after JSON unmarshaling into PapiEdgeHostnames
//
// See: edgegrid/json.Unmarshal()
func (availableBehaviors *PapiAvailableBehaviors) PostUnmarshalJSON() error {
	availableBehaviors.Init()

	for key := range availableBehaviors.Behaviors.Items {
		availableBehaviors.Behaviors.Items[key].parent = availableBehaviors
	}

	availableBehaviors.Complete <- true

	return nil
}

// GetAvailableBehaviors retrieves available behaviors for a given property
//
// See: PapiProperty.GetAvailableBehaviors
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listavailablebehaviors
// Endpoint: GET /papi/v0/properties/{propertyId}/versions/{propertyVersion}/available-behaviors{?contractId,groupId}
func (availableBehaviors *PapiAvailableBehaviors) GetAvailableBehaviors(property *PapiProperty) error {
	res, err := availableBehaviors.service.client.Get(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions/%d/available-behaviors?contractId=%s&groupId=%s",
			property.PropertyID,
			property.LatestVersion,
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

	newAvailableBehaviors := NewPapiAvailableBehaviors(availableBehaviors.service)
	if err = res.BodyJSON(newAvailableBehaviors); err != nil {
		return err
	}

	*availableBehaviors = *newAvailableBehaviors

	return nil
}

// PapiAvailableBehavior represents an available behavior resource
type PapiAvailableBehavior struct {
	resource
	parent     *PapiAvailableBehaviors
	Name       string `json:"name"`
	SchemaLink string `json:"schemaLink"`
}

// NewPapiAvailableBehavior creates a new PapiAvailableBehavior
func NewPapiAvailableBehavior(parent *PapiAvailableBehaviors) *PapiAvailableBehavior {
	availableBehavior := &PapiAvailableBehavior{parent: parent}
	availableBehavior.Init()

	return availableBehavior
}

// GetSchema retrieves the JSON schema for an available behavior
func (behavior *PapiAvailableBehavior) GetSchema() (*gojsonschema.Schema, error) {
	res, err := behavior.parent.service.client.Get(behavior.SchemaLink)

	if err != nil {
		return nil, err
	}

	schemaBytes, _ := ioutil.ReadAll(res.Body)
	schemaBody := string(schemaBytes)
	loader := gojsonschema.NewStringLoader(schemaBody)
	schema, err := gojsonschema.NewSchema(loader)

	return schema, err
}

// PapiRuleErrors represents an validate error returned for a rule
type PapiRuleErrors struct {
	resource
	Type         string `json:"type"`
	Title        string `json:"title"`
	Detail       string `json:"detail"`
	Instance     string `json:"instance"`
	BehaviorName string `json:"behaviorName"`
}

// NewPapiRuleErrors creates a new PapiRuleErrors
func NewPapiRuleErrors() *PapiRuleErrors {
	ruleErrors := &PapiRuleErrors{}
	ruleErrors.Init()

	return ruleErrors
}
