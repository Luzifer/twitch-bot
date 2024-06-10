package linkcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveMetaRedir(t *testing.T) {
	testDoc := []byte(`<!DOCTYPE html>
<html>
<head>
<title></title>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<meta property="twitter:image" content="">
<meta http-equiv='refresh' content='0; url=https://github.com/Luzifer/twitch-bot'>
</head>
<body>
</body>
</html>`)

	redir, err := resolveMetaRedirect(testDoc)
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/Luzifer/twitch-bot", redir)

	testDoc = []byte(`<!DOCTYPE html>
<html>
<head>
<title></title>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<meta property="twitter:image" content="">
</head>
<body>
</body>
</html>`)

	redir, err = resolveMetaRedirect(testDoc)
	require.ErrorIs(t, err, errNoMetaRedir)
	assert.Equal(t, "", redir)
}
