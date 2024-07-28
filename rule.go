package main

type Rule struct {
	Name     string `json:"name"`
	Match    Match  `json:"match"`
	Workflow string `json:"workflow"`
}

type Match struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RuleMatcher interface {
	Match(alert Alert) []string
}

type BasicRuleMatcher struct {
	rules []Rule
}

// Verify interface compliance
var _ RuleMatcher = (*BasicRuleMatcher)(nil)

func NewBasicRuleMatcher(rules []Rule) *BasicRuleMatcher {
	return &BasicRuleMatcher{
		rules: rules,
	}
}

func (r *BasicRuleMatcher) Match(alert Alert) (workflows []string) {
	// for dedupe
	matched := make(map[string]struct{})
	for _, rule := range r.rules {
		if _, ok := alert[rule.Match.Key]; ok && alert[rule.Match.Key] == rule.Match.Value {
			matched[rule.Workflow] = struct{}{}
		}
	}

	for wf := range matched {
		workflows = append(workflows, wf)
	}

	return
}

func LoadRules() ([]Rule, error) {
	var rules []Rule
	err := ParseJsonFileInto("config/rules.json", &rules)
	return rules, err
}
