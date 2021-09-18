package plugins

type RuleAction struct {
	Type       string          `json:"type" yaml:"type"`
	Attributes FieldCollection `json:"attributes" yaml:"attributes"`
}
