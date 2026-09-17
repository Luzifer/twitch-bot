package docgen

import (
	"testing"
	"time"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func TestRenderTplDocsExample(t *testing.T) {
	example := &plugins.TemplateFuncDocumentationExample{
		MatchMessage:   "!command (.*)",
		MessageContent: "!command value",
		Template:       "{{ .testDuration }}",
	}

	output, err := renderTplDocsExample(example, func(tpl string, msg *irc.Message, rule *plugins.Rule, fields *fieldcollection.FieldCollection) (string, error) {
		assert.Equal(t, example.Template, tpl)
		assert.Equal(t, example.MessageContent, msg.Trailing())
		assert.Equal(t, "exampleuser", msg.Name)
		require.NotNil(t, rule.MatchMessage)
		assert.Equal(t, example.MatchMessage, *rule.MatchMessage)
		assert.Equal(t, 5*time.Hour+33*time.Minute+12*time.Second, fields.MustDuration("testDuration", nil))

		return "rendered", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "rendered", output)
}
