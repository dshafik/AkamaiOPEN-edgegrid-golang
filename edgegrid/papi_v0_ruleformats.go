package edgegrid

type PapiRuleFormats struct {
	RuleFormats struct {
		Items []string `json:"items"`
	} `json:"ruleFormats"`
	Complete chan bool `json:"-"`
}

func NewPapiRuleFormats() *PapiRuleFormats {
	ruleFormats := &PapiRuleFormats{}
	ruleFormats.Init()

	return ruleFormats
}

func (ruleFormats *PapiRuleFormats) Init() {
	ruleFormats.Complete = make(chan bool, 1)
}

func (ruleFormats *PapiRuleFormats) PostUnmashalJSON() error {
	ruleFormats.Init()
	ruleFormats.Complete <- true

	return nil
}
