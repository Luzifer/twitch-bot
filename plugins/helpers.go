package plugins

import (
	"fmt"
	"strings"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"
)

// DeriveChannel takes an irc.Message and a FieldCollection and tries
// to extract from them the channel the event / message has taken place
func DeriveChannel(m *irc.Message, evtData *fieldcollection.FieldCollection) string {
	if m != nil && len(m.Params) > 0 && strings.HasPrefix(m.Params[0], "#") {
		return m.Params[0]
	}

	if s, err := evtData.String("channel"); err == nil {
		return fmt.Sprintf("#%s", strings.TrimLeft(s, "#"))
	}

	return ""
}

// DeriveUser takes an irc.Message and a FieldCollection and tries
// to extract from them the user causing the event / message
func DeriveUser(m *irc.Message, evtData *fieldcollection.FieldCollection) string {
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
