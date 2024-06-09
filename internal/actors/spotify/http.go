package spotify

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/oauth2"
)

const (
	spotifyRequestTimeout = 2 * time.Second

	pkcePBKDFIter = 210000
	pkcePBKDFLen  = 64
)

var instanceSalt = uuid.Must(uuid.NewV4()).String()

func handleStartAuth(w http.ResponseWriter, r *http.Request) {
	channel := mux.Vars(r)["channel"]
	pkceVerifier := hex.EncodeToString(pbkdf2.Key([]byte(channel), []byte(instanceSalt), pkcePBKDFIter, pkcePBKDFLen, sha512.New))

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
			conf.AuthCodeURL(
				fmt.Sprintf("%x", sha256.Sum256(append([]byte(conf.ClientID), []byte(channel)...))),
				oauth2.S256ChallengeOption(pkceVerifier),
			),
			http.StatusFound,
		)
		return
	}

	token, err := conf.Exchange(
		r.Context(),
		r.URL.Query().Get("code"),
		oauth2.VerifierOption(pkceVerifier),
	)
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

	http.Error(w, "Spotify is now authorized for this channel, you can close this page", http.StatusOK)
}
