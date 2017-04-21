package edgegrid

type PapiRuleFormats struct {
	resource
	RuleFormats struct {
		Items []string `json:"items"`
	} `json:"ruleFormats"`
}

func NewPapiRuleFormats() *PapiRuleFormats {
	ruleFormats := &PapiRuleFormats{}
	ruleFormats.Init()

	return ruleFormats
}
