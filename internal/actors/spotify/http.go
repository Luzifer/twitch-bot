package spotify

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const spotifyRequestTimeout = 2 * time.Second

func handleStartAuth(w http.ResponseWriter, r *http.Request) {
	channel := mux.Vars(r)["channel"]

	redirURL := baseURL.ResolveReference(&url.URL{Path: r.URL.Path})
	conf, err := oauthConfig(channel, strings.Split(redirURL.String(), "?")[0])
	if err != nil {
		logrus.WithError(err).Error("getting Spotify oauth config")
		http.Error(w, "unable to get Spotify config for this channel", http.StatusInternalServerError)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(
			w, r,
			conf.AuthCodeURL(fmt.Sprintf("%x", sha256.Sum256(append([]byte(conf.ClientID), []byte(channel)...)))),
			http.StatusFound,
		)
		return
	}

	token, err := conf.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		logrus.WithError(err).Error("getting Spotify oauth token")
		http.Error(w, "unable to get Spotify auth token", http.StatusInternalServerError)
		return
	}

	if err = db.StoreEncryptedCoreMeta(strings.Join([]string{"spotify-auth", channel}, ":"), token); err != nil {
		logrus.WithError(err).Error("storing Spotify oauth token")
		http.Error(w, "unable to store Spotify auth token", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Spotify is now authorized for this channel, you can close this page")
}

func oauthConfig(channel, redirectURL string) (conf *oauth2.Config, err error) {
	clientID, err := getModuleConfig(actorName, channel).String("clientId")
	if err != nil {
		return nil, fmt.Errorf("getting clientId for channel: %w", err)
	}

	clientSecret, err := getModuleConfig(actorName, channel).String("clientSecret")
	if err != nil {
		return nil, fmt.Errorf("getting clientSecret for channel: %w", err)
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
		RedirectURL: redirectURL,
		Scopes:      []string{"user-read-currently-playing"},
	}, nil
}
