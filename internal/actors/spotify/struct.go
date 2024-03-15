package spotify

type (
	currentPlayingTrackResponse struct {
		Device struct {
			ID               string `json:"id"`
			IsActive         bool   `json:"is_active"`
			IsPrivateSession bool   `json:"is_private_session"`
			IsRestricted     bool   `json:"is_restricted"`
			Name             string `json:"name"`
			Type             string `json:"type"`
			VolumePercent    int    `json:"volume_percent"`
			SupportsVolume   bool   `json:"supports_volume"`
		} `json:"device"`
		RepeatState  string `json:"repeat_state"`
		ShuffleState bool   `json:"shuffle_state"`
		Context      struct {
			Type         string `json:"type"`
			Href         string `json:"href"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			URI string `json:"uri"`
		} `json:"context"`
		Timestamp  int  `json:"timestamp"`
		ProgressMs int  `json:"progress_ms"`
		IsPlaying  bool `json:"is_playing"`
		Item       struct {
			Album struct {
				AlbumType        string   `json:"album_type"`
				TotalTracks      int      `json:"total_tracks"`
				AvailableMarkets []string `json:"available_markets"`
				ExternalUrls     struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href   string `json:"href"`
				ID     string `json:"id"`
				Images []struct {
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				} `json:"images"`
				Name                 string `json:"name"`
				ReleaseDate          string `json:"release_date"`
				ReleaseDatePrecision string `json:"release_date_precision"`
				Restrictions         struct {
					Reason string `json:"reason"`
				} `json:"restrictions"`
				Type    string `json:"type"`
				URI     string `json:"uri"`
				Artists []struct {
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"external_urls"`
					Href string `json:"href"`
					ID   string `json:"id"`
					Name string `json:"name"`
					Type string `json:"type"`
					URI  string `json:"uri"`
				} `json:"artists"`
			} `json:"album"`
			Artists []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Followers struct {
					Href  string `json:"href"`
					Total int    `json:"total"`
				} `json:"followers"`
				Genres []string `json:"genres"`
				Href   string   `json:"href"`
				ID     string   `json:"id"`
				Images []struct {
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				} `json:"images"`
				Name       string `json:"name"`
				Popularity int    `json:"popularity"`
				Type       string `json:"type"`
				URI        string `json:"uri"`
			} `json:"artists"`
			AvailableMarkets []string `json:"available_markets"`
			DiscNumber       int      `json:"disc_number"`
			DurationMs       int      `json:"duration_ms"`
			Explicit         bool     `json:"explicit"`
			ExternalIDs      struct {
				Isrc string `json:"isrc"`
				Ean  string `json:"ean"`
				Upc  string `json:"upc"`
			} `json:"external_ids"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href         string   `json:"href"`
			ID           string   `json:"id"`
			IsPlayable   bool     `json:"is_playable"`
			LinkedFrom   struct{} `json:"linked_from"`
			Restrictions struct {
				Reason string `json:"reason"`
			} `json:"restrictions"`
			Name        string `json:"name"`
			Popularity  int    `json:"popularity"`
			PreviewURL  string `json:"preview_url"`
			TrackNumber int    `json:"track_number"`
			Type        string `json:"type"`
			URI         string `json:"uri"`
			IsLocal     bool   `json:"is_local"`
		} `json:"item"`
		CurrentlyPlayingType string `json:"currently_playing_type"`
		Actions              struct {
			InterruptingPlayback  bool `json:"interrupting_playback"`
			Pausing               bool `json:"pausing"`
			Resuming              bool `json:"resuming"`
			Seeking               bool `json:"seeking"`
			SkippingNext          bool `json:"skipping_next"`
			SkippingPrev          bool `json:"skipping_prev"`
			TogglingRepeatContext bool `json:"toggling_repeat_context"`
			TogglingShuffle       bool `json:"toggling_shuffle"`
			TogglingRepeatTrack   bool `json:"toggling_repeat_track"`
			TransferringPlayback  bool `json:"transferring_playback"`
		} `json:"actions"`
	}
)
