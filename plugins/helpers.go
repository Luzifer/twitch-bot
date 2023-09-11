package plugins

import (
	"fmt"
	"strings"

	"gopkg.in/irc.v4"
)

func DeriveChannel(m *irc.Message, evtData *FieldCollection) string {
	if m != nil && len(m.Params) > 0 && strings.HasPrefix(m.Params[0], "#") {
		return m.Params[0]
	}

	if s, err := evtData.String("channel"); err == nil {
		return fmt.Sprintf("#%s", strings.TrimLeft(s, "#"))
	}

	return ""
}

func DeriveUser(m *irc.Message, evtData *FieldCollection) string {
	if m != nil && m.User != "" {
		return m.User
	}

	if s, err := evtData.String("user"); err == nil {
		return s
	}

	if s, err := evtData.String("username"); err == nil {
		return s
	}

	return ""
}
