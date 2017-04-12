package edgegrid

import (
	"encoding/json"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"strings"
)

type PapiRules struct {
	service         *PapiV0Service
	AccountId       string    `json:"accountId"`
	ContractId      string    `json:"contractId"`
	GroupId         string    `json:"groupId"`
	PropertyId      string    `json:"propertyId"`
	PropertyVersion int       `json:"propertyVersion"`
	Etag            string    `json:"etag"`
	RuleFormat      string    `json:"ruleFormat"`
	Rules           *PapiRule `json:"rules"`
}

func (rules *PapiRules) UnmarshalJSON(b []byte) error {
	type PapiRulesTemp PapiRules
	temp := &PapiRulesTemp{service: rules.service}

	fmt.Println(string(b))

	if err := json.Unmarshal(b, temp); err != nil {
		return err
	}
	*rules = (PapiRules)(*temp)

	for key, _ := range rules.Rules.Behaviors {
		rules.Rules.Behaviors[key].parent = rules.Rules
		if len(rules.Rules.Children) > 0 {
			for _, v := range rules.Rules.getChildren(0) {
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
		children := rules.Rules.getChildren(0)
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

type PapiRule struct {
	parent  *PapiRules
	Name    string `json:"name"`
	Options struct {
		IsSecure bool `json:"is_secure,omitempty"`
	} `json:"options,omitempty"`
	Criteria       []*PapiCriteria  `json:"criteria,omitempty"`
	Behaviors      []*PapiBehavior  `json:"behaviors"`
	Children       []*PapiChildRule `json:"children,omitempty"`
	Comment        string           `json:"comment,omitempty"`
	CriteriaLocked bool             `json:criteriaLocked,omitempty`
}

func (rule *PapiRule) getChildren(depth int) []*PapiChildRule {
	depth += 1
	var children []*PapiChildRule
	if len(rule.Children) > 0 {
		for _, v := range rule.Children {
			v.depth = depth
			children = append(children, v)
			children = append(children, v.getChildren(depth)...)
		}
	}

	return children
}

type PapiChildRule struct {
	PapiRule
	depth     int
	Behaviors []*PapiBehavior `json:"behaviors,omitempty"`
}

type PapiCriteria struct {
	parent  *PapiRule
	Name    string           `json:"name"`
	Options *PapiOptionValue `json:"options"`
}

func (criteria *PapiCriteria) validateOptions() error {
	return nil
}

type PapiBehavior struct {
	parent  *PapiRule
	Name    string           `json:"name"`
	Options *PapiOptionValue `json:"options"`
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
}

func (behaviors *PapiAvailableBehaviors) UnmarshalJSON(b []byte) error {
	type PapiAvailableBehaviorsTemp PapiAvailableBehaviors
	temp := &PapiAvailableBehaviorsTemp{service: behaviors.service}

	if err := json.Unmarshal(b, temp); err != nil {
		return err
	}
	*behaviors = (PapiAvailableBehaviors)(*temp)

	for key, _ := range behaviors.Behaviors.Items {
		behaviors.Behaviors.Items[key].parent = behaviors
	}

	return nil
}

type PapiAvailableBehavior struct {
	parent     *PapiAvailableBehaviors
	Name       string `json:"name"`
	SchemaLink string `json:"schemaLink"`
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
