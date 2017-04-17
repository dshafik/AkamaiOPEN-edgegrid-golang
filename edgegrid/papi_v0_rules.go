package edgegrid

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"strings"
)

type PapiRules struct {
	service         *PapiV0Service
	AccountId       string            `json:"accountId"`
	ContractId      string            `json:"contractId"`
	GroupId         string            `json:"groupId"`
	PropertyId      string            `json:"propertyId"`
	PropertyVersion int               `json:"propertyVersion"`
	Etag            string            `json:"etag"`
	RuleFormat      string            `json:"ruleFormat"`
	Rules           *PapiRule         `json:"rules"`
	Errors          []*PapiRuleErrors `json:"errors,omitempty"`
	Complete        chan bool         `json:"-"`
}

func NewPapiRules(service *PapiV0Service) *PapiRules {
	rules := &PapiRules{service: service}
	rules.Init()

	return rules
}

func (rules *PapiRules) Init() {
	rules.Complete = make(chan bool, 1)
}

func (rules *PapiRules) PostUnmarshalJSON() error {
	rules.Init()

	for key, _ := range rules.Rules.Behaviors {
		rules.Rules.Behaviors[key].parent = rules.Rules
		if len(rules.Rules.Children) > 0 {
			for _, v := range rules.Rules.GetChildren(0, 0) {
				for _, j := range v.Behaviors {
					j.parent = rules.Rules
				}
			}
		}
	}

	for key, _ := range rules.Rules.Criteria {
		rules.Rules.Criteria[key].parent = rules.Rules
	}

	return nil
}

func (rules *PapiRules) PreMarshalJSON() error {
	rules.Errors = nil
	return nil
}

func (rules *PapiRules) PrintRules() error {
	groups := &PapiGroup{
		parent:      &PapiGroups{service: rules.service},
		GroupId:     rules.GroupId,
		ContractIds: []string{rules.ContractId},
	}

	properties, _ := groups.GetProperties(nil)
	var property *PapiProperty
	for _, property = range properties.Properties.Items {
		if property.PropertyId == rules.PropertyId {
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
			spacer := strings.TrimSuffix(strings.Repeat(prefix, child.depth), "│")
			if i < len(children) {
				fmt.Printf("%s├── Section: %s\n", spacer, child.Name)
			} else {
				fmt.Printf("%s└── Section: %s\n", spacer, child.Name)
			}

			spacer = strings.TrimSuffix(strings.Repeat(prefix, child.depth+1), "│")
			j := 0
			for _, behavior := range child.Behaviors {
				j++
				if j < len(child.Behaviors) {
					fmt.Printf("%s├── Behavior: %s\n", spacer, behavior.Name)
				} else {
					//spacer = strings.TrimSuffix(spacer, "│   ") + "    "
					fmt.Printf("%s└── Behavior: %s\n", spacer, behavior.Name)
				}
				space := strings.TrimSuffix(strings.Repeat(prefix, child.depth+2), "│")

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

func (rules *PapiRules) GetRules() []*PapiRule {
	var flatRules []*PapiRule
	flatRules = append(flatRules, rules.Rules)
	for _, rules := range rules.Rules.GetChildren(0, 0) {
		flatRules = append(flatRules, rules)
	}

	return flatRules
}

func (rules *PapiRules) Save() error {
	// /papi/v0/properties/{propertyId}/versions/{propertyVersion}/rules/{?contractId,groupId}
	res, err := rules.service.client.PutJson(
		fmt.Sprintf(
			"/papi/v0/properties/%s/versions/%d/rules/?contractId=%s&groupId=%s",
			rules.PropertyId,
			rules.PropertyVersion,
			rules.ContractId,
			rules.GroupId,
		),
		rules,
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return NewApiError(res)
	}

	res.BodyJson(rules)

	if len(rules.Errors) != 0 {
		return fmt.Errorf("There were %d errors. See rules.Errors for details.", len(rules.Errors))
	}

	return nil
}

func (rules *PapiRules) SetBehaviorOptions(path string, newOptions PapiOptionValue) error {
	behavior, err := rules.FindBehavior(path)
	if err != nil {
		return err
	}

	behavior.Options = &newOptions
	return nil
}

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

type PapiRule struct {
	parent  *PapiRules
	depth   int
	Name    string `json:"name"`
	Options struct {
		IsSecure bool `json:"is_secure,omitempty"`
	} `json:"options,omitempty"`
	Criteria       []*PapiCriteria `json:"criteria,omitempty"`
	Behaviors      []*PapiBehavior `json:"behaviors,omitempty"`
	Children       []*PapiRule     `json:"children,omitempty"`
	Comment        string          `json:"comment,omitempty"`
	CriteriaLocked bool            `json:criteriaLocked,omitempty`
	Complete       chan bool       `json:"-"`
}

func NewPapiRule(parent *PapiRules) *PapiRule {
	rule := &PapiRule{parent: parent}
	rule.Init()

	return rule
}

func (rule *PapiRule) Init() {
	rule.Complete = make(chan bool, 1)
}

func (rule *PapiRule) PostUnmashalJSON() error {
	rule.Init()
	rule.Complete <- true

	return nil
}

func (rule *PapiRule) GetChildren(depth int, limit int) []*PapiRule {
	depth += 1

	if limit != 0 && depth > limit {
		return nil
	}

	var children []*PapiRule
	if len(rule.Children) > 0 {
		for _, v := range rule.Children {
			v.depth = depth
			children = append(children, v)
			children = append(children, v.GetChildren(depth, limit)...)
		}
	}

	return children
}

func (rule *PapiRule) AddChildRule(child *PapiRule) {
	rule.Children = append(rule.Children, child)
}

func (rule *PapiRule) AddCriteria(critera *PapiCriteria) {
	rule.Criteria = append(rule.Criteria, critera)
}

func (rule *PapiRule) AddBehavior(behavior *PapiBehavior) {
	rule.Behaviors = append(rule.Behaviors, behavior)
}

type PapiCriteria struct {
	parent   *PapiRule
	Name     string           `json:"name"`
	Options  *PapiOptionValue `json:"options"`
	Complete chan bool        `json:"-"`
}

func NewPapiCriteria(parent *PapiRule) *PapiCriteria {
	criteria := &PapiCriteria{parent: parent}
	criteria.Init()

	return criteria
}

func (criteria *PapiCriteria) Init() {
	criteria.Complete = make(chan bool, 1)
}

func (criteria *PapiCriteria) PostUnmashalJSON() error {
	criteria.Init()
	criteria.Complete <- true

	return nil
}

func (criteria *PapiCriteria) validateOptions() error {
	return nil
}

type PapiBehavior struct {
	parent   *PapiRule
	Name     string           `json:"name"`
	Options  *PapiOptionValue `json:"options"`
	Complete chan bool        `json:"-"`
}

func NewPapiBehavior(parent *PapiRule) *PapiBehavior {
	behavior := &PapiBehavior{parent: parent}
	behavior.Init()

	return behavior
}

func (behavior *PapiBehavior) Init() {
	behavior.Complete = make(chan bool, 1)
}

func (behavior *PapiBehavior) PostUnmashalJSON() error {
	behavior.Init()
	behavior.Complete <- true

	return nil
}

func (behavior *PapiBehavior) validateOptions() error {
	return nil
}

type PapiOptionValue map[string]interface{}

type PapiAvailableCriteria struct {
	service           *PapiV0Service
	ContractId        string `json:"contractId"`
	GroupId           string `json:"groupId"`
	ProductId         string `json:"productId"`
	RuleFormat        string `json:"ruleFormat"`
	AvailableCriteria struct {
		Items []struct {
			Name       string `json:"name"`
			SchemaLink string `json:"schemaLink"`
		} `json:"items"`
	} `json:"availableCriteria"`
	Complete chan bool `json:"-"`
}

func NewPapiAvailableCriteria(service *PapiV0Service) *PapiAvailableCriteria {
	availableCriteria := &PapiAvailableCriteria{service: service}
	availableCriteria.Init()

	return availableCriteria
}

func (availableCriteria *PapiAvailableCriteria) Init() {
	availableCriteria.Complete = make(chan bool, 1)
}

func (availableCriteria *PapiAvailableCriteria) PostUnmashalJSON() error {
	availableCriteria.Init()
	availableCriteria.Complete <- true

	return nil
}

type PapiAvailableBehaviors struct {
	service    *PapiV0Service
	ContractId string `json:"contractId"`
	GroupId    string `json:"groupId"`
	ProductId  string `json:"productId"`
	RuleFormat string `json:"ruleFormat"`
	Behaviors  struct {
		Items []PapiAvailableBehavior `json:"items"`
	} `json:"behaviors"`
	Complete chan bool `json:"-"`
}

func NewPapiAvailableBehaviors(service *PapiV0Service) *PapiAvailableBehaviors {
	availableBehaviors := &PapiAvailableBehaviors{service: service}
	availableBehaviors.Init()

	return availableBehaviors
}

func (availableBehaviors *PapiAvailableBehaviors) Init() {
	availableBehaviors.Complete = make(chan bool, 1)
}

func (availableBehaviors *PapiAvailableBehaviors) PostUnmashalJSON() error {
	availableBehaviors.Init()

	for key, _ := range availableBehaviors.Behaviors.Items {
		availableBehaviors.Behaviors.Items[key].parent = availableBehaviors
	}

	return nil

	availableBehaviors.Complete <- true

	return nil
}

type PapiAvailableBehavior struct {
	parent     *PapiAvailableBehaviors
	Name       string    `json:"name"`
	SchemaLink string    `json:"schemaLink"`
	Complete   chan bool `json:"-"`
}

func NewPapiAvailableBehavior(parent *PapiAvailableBehaviors) *PapiAvailableBehavior {
	availableBehavior := &PapiAvailableBehavior{parent: parent}
	availableBehavior.Init()

	return availableBehavior
}

func (availableBehavior *PapiAvailableBehavior) Init() {
	availableBehavior.Complete = make(chan bool, 1)
}

func (availableBehavior *PapiAvailableBehavior) PostUnmashalJSON() error {
	availableBehavior.Init()
	availableBehavior.Complete <- true

	return nil
}

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

type PapiRuleErrors struct {
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Detail       string    `json:"detail"`
	Instance     string    `json:"instance"`
	BehaviorName string    `json:"behaviorName"`
	Complete     chan bool `json:"-"`
}

func NewPapiRuleErrors() *PapiRuleErrors {
	ruleErrors := &PapiRuleErrors{}
	ruleErrors.Init()

	return ruleErrors
}

func (ruleErrors *PapiRuleErrors) Init() {
	ruleErrors.Complete = make(chan bool, 1)
}

func (ruleErrors *PapiRuleErrors) PostUnmashalJSON() error {
	ruleErrors.Init()
	ruleErrors.Complete <- true

	return nil
}
