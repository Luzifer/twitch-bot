package plugins

type RuleAction struct {
	Type       string               `json:"type" yaml:"type"`
	Attributes moduleAttributeStore `json:"attributes" yaml:"attributes"`
}
