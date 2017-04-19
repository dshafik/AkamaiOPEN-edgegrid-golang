package edgegrid

type PapiRuleFormats struct {
	Resource
	RuleFormats struct {
		Items []string `json:"items"`
	} `json:"ruleFormats"`
}

func NewPapiRuleFormats() *PapiRuleFormats {
	ruleFormats := &PapiRuleFormats{}
	ruleFormats.Init()

	return ruleFormats
}
