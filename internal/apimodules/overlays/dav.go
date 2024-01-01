package overlays

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/webdav"
)

func getDAVHandler() http.HandlerFunc {
	overlaysDir := os.Getenv("OVERLAYS_DIR")
	if ds, err := os.Stat(overlaysDir); err != nil || overlaysDir == "" || !ds.IsDir() {
		return http.NotFound
	}

	return (&webdav.Handler{
		Prefix:     "/overlays/dav",
		FileSystem: webdav.Dir(overlaysDir),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			logger := logrus.WithField("module", "overlays-dav")
			if err != nil {
				logger = logger.WithError(err)
			}
			logger.Debugf("%s %s", r.Method, r.URL)
		},
	}).ServeHTTP
}
