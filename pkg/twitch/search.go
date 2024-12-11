package twitch

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type (
	// Category represents information about a category
	Category struct {
		BoxArtURL string `json:"box_art_url"`
		ID        string `json:"id"`
		Name      string `json:"name"`
	}
)

// SearchCategories returns a list of categories matching the given
// search. The result is not exact but contains Twitchs fuzzy-results.
func (c *Client) SearchCategories(ctx context.Context, name string) ([]Category, error) {
	var out []Category

	params := make(url.Values)
	params.Set("query", name)
	params.Set("first", "100")

	var resp struct {
		Data       []Category `json:"data"`
		Pagination struct {
			Cursor string `json:"cursor"`
		} `json:"pagination"`
	}

	for {
		if err := c.Request(ctx, ClientRequestOpts{
			AuthType: AuthTypeAppAccessToken,
			Method:   http.MethodGet,
			OKStatus: http.StatusOK,
			Out:      &resp,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/search/categories?%s", params.Encode()),
		}); err != nil {
			return nil, errors.Wrap(err, "executing request")
		}

		out = append(out, resp.Data...)

		if resp.Pagination.Cursor == "" {
			break
		}

		params.Set("after", resp.Pagination.Cursor)
		resp.Pagination.Cursor = "" // Clear from struct as struct is reused
	}

	return out, nil
}
