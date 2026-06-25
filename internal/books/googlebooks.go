package books

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type GoogleBooksClient struct {
	HTTPClient *http.Client
	APIKey     string
}

type googleBooksSearchResponse struct {
	Items []googleBooksVolume `json:"items"`
}

type googleBooksVolume struct {
	ID         string                `json:"id"`
	VolumeInfo googleBooksVolumeInfo `json:"volumeInfo"`
}

type googleBooksVolumeInfo struct {
	Title               string                          `json:"title"`
	Authors             []string                        `json:"authors"`
	Description         string                          `json:"description"`
	PublishedDate       string                          `json:"publishedDate"`
	Categories          []string                        `json:"categories"`
	Language            string                          `json:"language"`
	ImageLinks          googleBooksImageLinks           `json:"imageLinks"`
	IndustryIdentifiers []googleBooksIndustryIdentifier `json:"industryIdentifiers"`
}

type googleBooksImageLinks struct {
	SmallThumbnail string `json:"smallThumbnail"`
	Thumbnail      string `json:"thumbnail"`
}

type googleBooksIndustryIdentifier struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

func (c GoogleBooksClient) SearchBookMetadataCandidates(ctx context.Context, title, author string) ([]BookMetadataCandidate, error) {
	if strings.TrimSpace(title) == "" {
		return nil, nil
	}

	u, err := url.Parse("https://www.googleapis.com/books/v1/volumes")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	queryParts := []string{"intitle:" + title}
	if strings.TrimSpace(author) != "" {
		queryParts = append(queryParts, "inauthor:"+author)
	}
	q.Set("q", strings.Join(queryParts, " "))
	q.Set("maxResults", "5")
	q.Set("printType", "books")
	q.Set("fields", "items(id,volumeInfo(title,authors,description,publishedDate,categories,language,imageLinks(thumbnail,smallThumbnail),industryIdentifiers(type,identifier)))")
	if strings.TrimSpace(c.APIKey) != "" {
		q.Set("key", c.APIKey)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Librorum/0.1")

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
		return nil, fmt.Errorf("google books search failed: %s", resp.Status)
	}

	var payload googleBooksSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	candidates := make([]BookMetadataCandidate, 0, len(payload.Items))
	for _, item := range payload.Items {
		info := item.VolumeInfo
		if strings.TrimSpace(info.Title) == "" {
			continue
		}

		isbns := googleBooksISBNs(info.IndustryIdentifiers)
		candidates = append(candidates, BookMetadataCandidate{
			Source:          MetadataSourceGoogleBooks,
			SourceID:        item.ID,
			Title:           strings.TrimSpace(info.Title),
			Author:          strings.Join(nonEmptyStrings(info.Authors), ", "),
			Description:     html.UnescapeString(strings.TrimSpace(info.Description)),
			Genres:          nonNilStrings(NormalizeGenres(limitStrings(info.Categories, 10))),
			Language:        strings.TrimSpace(info.Language),
			PublicationYear: googleBooksPublicationYear(info.PublishedDate),
			ISBN:            firstString(isbns),
			ISBNs:           isbns,
			CoverURL:        normalizeGoogleBooksImageURL(firstString([]string{info.ImageLinks.Thumbnail, info.ImageLinks.SmallThumbnail})),
		})
	}
	return candidates, nil
}

func googleBooksISBNs(identifiers []googleBooksIndustryIdentifier) []string {
	ordered := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		if strings.EqualFold(identifier.Type, "ISBN_13") {
			ordered = append(ordered, identifier.Identifier)
		}
	}
	for _, identifier := range identifiers {
		if strings.EqualFold(identifier.Type, "ISBN_10") {
			ordered = append(ordered, identifier.Identifier)
		}
	}
	return uniqueISBNs(ordered)
}

func googleBooksPublicationYear(publishedDate string) *int32 {
	publishedDate = strings.TrimSpace(publishedDate)
	if len(publishedDate) < 4 {
		return nil
	}
	year, err := strconv.ParseInt(publishedDate[:4], 10, 32)
	if err != nil || year <= 0 {
		return nil
	}
	publicationYear := int32(year)
	return &publicationYear
}

func normalizeGoogleBooksImageURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	parsed, err := url.Parse(value)
	if err != nil {
		if strings.HasPrefix(value, "http://") {
			return "https://" + strings.TrimPrefix(value, "http://")
		}
		return value
	}
	if parsed.Scheme == "http" {
		parsed.Scheme = "https"
	}
	query := parsed.Query()
	query.Del("edge")
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func nonEmptyStrings(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized
}
