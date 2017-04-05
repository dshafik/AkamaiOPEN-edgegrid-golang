package edgegrid

import (
	"encoding/json"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
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
			for _, v := range rules.Rules.getChildren() {
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

func (rule *PapiRule) getChildren() []*PapiChildRule {
	var children []*PapiChildRule
	if len(rule.Children) > 0 {
		children = rule.Children
		for _, v := range rule.Children {
			children = append(children, v.getChildren()...)
		}
	}

	return children
}

type PapiChildRule struct {
	PapiRule
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
