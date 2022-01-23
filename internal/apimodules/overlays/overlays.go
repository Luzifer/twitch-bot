package overlays

import (
	"embed"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/plugins"
)

var (
	//go:embed default/**
	embeddedOverlays embed.FS

	fsStack httpFSStack
)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
		HandlerFunc:  handleServeOverlayAsset,
		IsPrefix:     true,
		Method:       http.MethodGet,
		Module:       "overlays",
		Path:         "/",
		ResponseType: plugins.HTTPRouteResponseTypeMultiple,
	})

	fsStack = httpFSStack{
		newPrefixedFS("default", http.FS(embeddedOverlays)),
	}

	overlaysDir := os.Getenv("OVERLAYS_DIR")
	if ds, err := os.Stat(overlaysDir); err != nil || overlaysDir == "" || !ds.IsDir() {
		log.WithField("dir", overlaysDir).Warn("Overlays dir not specified, no dir or non existent")
	} else {
		fsStack = append(httpFSStack{http.Dir(overlaysDir)}, fsStack...)
	}

	return nil
}

func handleServeOverlayAsset(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/overlays", http.FileServer(fsStack)).ServeHTTP(w, r)
}
