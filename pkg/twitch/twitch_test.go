package twitch

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactURLSecrets(t *testing.T) {
	client := New("myClientId", "myClient+Secret", "", "")

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := client.GetTwitchAppAccessToken(ctx)
	urlErr := new(url.Error)
	require.ErrorAs(t, err, &urlErr)

	assert.NotContains(t, urlErr.URL, client.clientSecret)
	assert.NotContains(t, urlErr.URL, url.QueryEscape(client.clientSecret))
	assert.Contains(t, urlErr.URL, client.hashSecret(client.clientSecret))
	assert.Contains(t, urlErr.URL, client.clientID)
}
