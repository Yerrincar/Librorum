package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OpenLibraryClient struct {
	HTTPClient *http.Client
	Contact    string
}

type openLibrarySearchResponse struct {
	NumFound int                  `json:"numFound"`
	Docs     []openLibraryBookDoc `json:"docs"`
}

type openLibraryBookDoc struct {
	CoverID int `json:"cover_i"`
}

func (c OpenLibraryClient) SearchCoverID(ctx context.Context, title, author string) (int, error) {
	if strings.TrimSpace(title) == "" {
		return 0, nil
	}

	u, err := url.Parse("https://openlibrary.org/search.json")
	if err != nil {
		return 0, err
	}
	q := u.Query()
	q.Set("title", title)
	if strings.TrimSpace(author) != "" {
		q.Set("author", author)
	}
	q.Set("limit", "5")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", c.UserAgent())

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("openlibrary search failed: %s", resp.Status)
	}

	var payload openLibrarySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}
	if payload.NumFound == 0 {
		return 0, nil
	}

	for _, doc := range payload.Docs {
		if doc.CoverID > 0 {
			return doc.CoverID, nil
		}
	}
	return 0, nil
}

func (c OpenLibraryClient) UserAgent() string {
	contact := strings.TrimSpace(c.Contact)
	if contact == "" {
		return "Librorum/0.1"
	}
	return "Librorum/0.1 (contact: " + contact + ")"
}
