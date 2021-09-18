package plugins

type RuleAction struct {
	Type       string         `json:"type" yaml:"type"`
	Attributes AttributeStore `json:"attributes" yaml:"attributes"`
}
