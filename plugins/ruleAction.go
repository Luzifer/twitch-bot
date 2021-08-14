package plugins

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
)

type RuleAction struct {
	yamlUnmarshal func(interface{}) error
	jsonValue     []byte
}

func (r *RuleAction) UnmarshalJSON(d []byte) error {
	r.jsonValue = d
	return nil
}

func (r *RuleAction) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r.yamlUnmarshal = unmarshal
	return nil
}

func (r *RuleAction) Unmarshal(v interface{}) error {
	switch {
	case r.yamlUnmarshal != nil:
		return r.yamlUnmarshal(v)

	case r.jsonValue != nil:
		jd := json.NewDecoder(bytes.NewReader(r.jsonValue))
		jd.DisallowUnknownFields()
		return jd.Decode(v)

	default:
		return errors.New("unmarshal on unprimed object")
	}
}
