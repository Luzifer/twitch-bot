package fixture

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func cookieGuardRedirect(cookieKey, cookieValue, target string) fixtureGeneratorFunc {
	return func(r *http.Request) (*http.Response, error) {
		cookie, err := r.Cookie(cookieKey)
		if errors.Is(err, http.ErrNoCookie) || cookie.Value != cookieValue {
			// No cookie, no arms! (Wait, what?)
			return &http.Response{
				Header: http.Header{
					"Set-Cookie": []string{(&http.Cookie{
						Name:     cookieKey,
						Value:    cookieValue,
						Path:     "/",
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteLaxMode,
					}).String()},
				},
				Request:    r,
				StatusCode: http.StatusOK,
			}, nil
		}

		// They had cookies! And they were tasty!
		return headerRedirect(http.StatusFound, target)(r)
	}
}

func headerRedirect(status int, target string) fixtureGeneratorFunc {
	return func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Header: http.Header{
				"Location": []string{target},
			},
			Request:    r,
			StatusCode: status,
		}, nil
	}
}

func metaRedirect(target string) fixtureGeneratorFunc {
	return func(r *http.Request) (*http.Response, error) {
		body := strings.TrimSpace(`
<!DOCTYPE html>
<html lang="en-US">
  <meta charset="utf-8">
  <title>Redirecting&hellip;</title>
  <link rel="canonical" href="%[1]s">
  <script>location="%[1]s"</script>
  <meta http-equiv="refresh" content="0; url=%[1]s">
  <meta name="robots" content="noindex">
  <h1>Redirecting&hellip;</h1>
  <a href="%[1]s">Click here if you are not redirected.</a>
</html>
		`)

		return &http.Response{
			Body:       io.NopCloser(strings.NewReader(fmt.Sprintf(body, target))),
			Request:    r,
			StatusCode: http.StatusOK,
		}, nil
	}
}

func redirectHTTPS(r *http.Request) (*http.Response, error) {
	u := r.URL.Clone()
	u.Scheme = "https"

	return headerRedirect(http.StatusFound, u.String())(r)
}
