package books

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
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	AuthorNames      []string `json:"author_name"`
	FirstPublishYear int32    `json:"first_publish_year"`
	Subjects         []string `json:"subject"`
	Languages        []string `json:"language"`
	CoverID          int      `json:"cover_i"`
	EditionKeys      []string `json:"edition_key"`
	ISBNs            []string `json:"isbn"`
}

type openLibraryWorkResponse struct {
	Description json.RawMessage `json:"description"`
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
	q.Set("limit", "20")
	q.Set("fields", "key,title,author_name,first_publish_year,subject,language,cover_i,edition_key,isbn")
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

func (c OpenLibraryClient) SearchBookMetadata(ctx context.Context, title, author string) (*BookMetadataCandidate, error) {
	candidates, err := c.SearchBookMetadataCandidates(ctx, title, author)
	if err != nil || len(candidates) == 0 {
		return nil, err
	}

	metadata := candidates[0]
	if description, err := c.WorkDescription(ctx, metadata.WorkKey); err == nil {
		metadata.Description = description
	}
	return &metadata, nil
}

func (c OpenLibraryClient) SearchBookMetadataCandidates(ctx context.Context, title, author string) ([]BookMetadataCandidate, error) {
	if strings.TrimSpace(title) == "" {
		return nil, nil
	}
	u, err := url.Parse("https://openlibrary.org/search.json")
	if err != nil {
		return nil, err
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
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent())

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openlibrary search failed: %s", resp.Status)
	}

	var payload openLibrarySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.NumFound == 0 {
		return nil, nil
	}

	candidates := make([]BookMetadataCandidate, 0, len(payload.Docs))
	for _, doc := range payload.Docs {
		if strings.TrimSpace(doc.Title) == "" {
			continue
		}

		isbns := uniqueISBNs(doc.ISBNs)
		metadata := BookMetadataCandidate{
			Source:   MetadataSourceOpenLibrary,
			SourceID: doc.Key,
			Title:    doc.Title,
			Genres:   nonNilStrings(NormalizeGenres(limitStrings(doc.Subjects, 10))),
			Language: firstString(doc.Languages),
			ISBN:     firstString(isbns),
			ISBNs:    isbns,
			CoverID:  doc.CoverID,
			WorkKey:  doc.Key,
		}
		if len(doc.AuthorNames) > 0 {
			metadata.Author = doc.AuthorNames[0]
		}
		if doc.FirstPublishYear > 0 {
			year := doc.FirstPublishYear
			metadata.PublicationYear = &year
		}
		candidates = append(candidates, metadata)
	}
	return candidates, nil
}

func (c OpenLibraryClient) WorkDescription(ctx context.Context, workKey string) (string, error) {
	workKey = strings.TrimSpace(workKey)
	if workKey == "" {
		return "", nil
	}

	workKey = strings.TrimSuffix(workKey, ".json")
	if !strings.HasPrefix(workKey, "/") {
		workKey = "/" + workKey
	}

	u, err := url.Parse("https://openlibrary.org" + workKey + ".json")
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", c.UserAgent())

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openlibrary work fetch failed: %s", resp.Status)
	}

	var payload openLibraryWorkResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	return parseOpenLibraryDescription(payload.Description), nil
}

func parseOpenLibraryDescription(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}

	var object struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(raw, &object); err == nil {
		return strings.TrimSpace(object.Value)
	}

	return ""
}

func firstString(values []string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func limitStrings(values []string, limit int) []string {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func (c OpenLibraryClient) UserAgent() string {
	contact := strings.TrimSpace(c.Contact)
	if contact == "" {
		return "Librorum/0.1"
	}
	return "Librorum/0.1 (contact: " + contact + ")"
}
