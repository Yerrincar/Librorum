package books

import (
	"Librorum/internal/storage"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type calibreOPFPackage struct {
	Metadata calibreOPFMetadata `xml:"metadata"`
}

type calibreOPFMetadata struct {
	Identifiers []calibreOPFIdentifier `xml:"http://purl.org/dc/elements/1.1/ identifier"`
	Title       string                 `xml:"http://purl.org/dc/elements/1.1/ title"`
	Creators    []string               `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Date        string                 `xml:"http://purl.org/dc/elements/1.1/ date"`
	Description string                 `xml:"http://purl.org/dc/elements/1.1/ description"`
	Languages   []string               `xml:"http://purl.org/dc/elements/1.1/ language"`
	Subjects    []string               `xml:"http://purl.org/dc/elements/1.1/ subject"`
}

type calibreOPFIdentifier struct {
	Value string     `xml:",chardata"`
	Attrs []xml.Attr `xml:",any,attr"`
}

var htmlTagRE = regexp.MustCompile(`<[^>]*>`)

type calibreRetryCandidate struct {
	Title string
	ISBN  string
}

func (h *Handler) CalibreMetadata(ctx context.Context, coverDir, dataDir, title, author string) (*BookMetadataCandidate, error) {
	title = strings.TrimSpace(title)
	author = strings.TrimSpace(author)
	if title == "" {
		return nil, nil
	}
	coverDir, err := filepath.Abs(coverDir)
	if err != nil {
		return nil, err
	}
	dataDir, err = filepath.Abs(dataDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(coverDir, 0o755); err != nil {
		return nil, err
	}

	tempRoot := filepath.Join(dataDir, "calibre")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		return nil, err
	}

	metadata, err := h.fetchCalibreMetadata(ctx, coverDir, tempRoot, title, author, "")
	if err != nil || metadata == nil || metadata.CoverPath != "" || metadata.ISBN == "" {
		if metadata != nil && metadata.CoverPath == "" {
			h.Logger.Info("Calibre returned metadata without a cover", map[string]string{"title": metadata.Title, "isbn": metadata.ISBN})
		}
		return metadata, err
	}

	isbnMetadata, err := h.fetchCalibreMetadata(ctx, coverDir, tempRoot, title, author, metadata.ISBN)
	if err != nil {
		h.Logger.Error("Error trying to fetch Calibre metadata by ISBN: "+err.Error(), map[string]string{"isbn": metadata.ISBN})
		return metadata, nil
	}
	if isbnMetadata == nil {
		return metadata, nil
	}
	if isbnMetadata.CoverPath == "" {
		h.Logger.Info("Calibre ISBN retry returned metadata without a cover", map[string]string{"title": isbnMetadata.Title, "isbn": metadata.ISBN})
		return metadata, nil
	}
	if !titleLikelyMatches(title, isbnMetadata.Title) {
		h.Logger.Info("Rejected Calibre ISBN retry cover for mismatched title", map[string]string{"requested_title": title, "returned_title": isbnMetadata.Title, "isbn": metadata.ISBN})
		_ = os.Remove(isbnMetadata.CoverPath)
		return metadata, nil
	}
	if isbnMetadata.ISBN == "" {
		isbnMetadata.ISBN = metadata.ISBN
	}
	return isbnMetadata, nil
}

func (h *Handler) RetryCalibreMetadataCover(ctx context.Context, coverDir, dataDir, title, author string, metadata *BookMetadataCandidate, retryCandidates []calibreRetryCandidate) (*BookMetadataCandidate, error) {
	if metadata == nil {
		return metadata, nil
	}

	title = strings.TrimSpace(title)
	author = strings.TrimSpace(author)
	if title == "" {
		return metadata, nil
	}

	metadataMatches := titleLikelyMatches(title, metadata.Title)
	if metadata.CoverPath != "" {
		if metadataMatches {
			return metadata, nil
		}
		h.Logger.Info("Rejected Calibre cover for mismatched title", map[string]string{"requested_title": title, "returned_title": metadata.Title, "cover_path": metadata.CoverPath})
		_ = os.Remove(metadata.CoverPath)
		metadata.CoverPath = ""
	}

	coverDir, err := filepath.Abs(coverDir)
	if err != nil {
		return metadata, err
	}
	dataDir, err = filepath.Abs(dataDir)
	if err != nil {
		return metadata, err
	}
	if err := os.MkdirAll(coverDir, 0o755); err != nil {
		return metadata, err
	}

	tempRoot := filepath.Join(dataDir, "calibre")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		return metadata, err
	}

	attempted := make(map[string]struct{})
	attempt := func(searchTitle, isbn string) *BookMetadataCandidate {
		searchTitle = strings.TrimSpace(searchTitle)
		isbn = normalizeISBN(isbn)
		if searchTitle == "" {
			return nil
		}
		if isbn == "" && strings.EqualFold(searchTitle, title) {
			return nil
		}

		key := strings.ToLower(searchTitle) + "\x00" + isbn
		if _, ok := attempted[key]; ok {
			return nil
		}
		attempted[key] = struct{}{}

		isbnMetadata, err := h.fetchCalibreMetadata(ctx, coverDir, tempRoot, searchTitle, author, isbn)
		if err != nil {
			h.Logger.Error("Error trying to fetch Calibre metadata by external ISBN: "+err.Error(), map[string]string{"isbn": isbn, "title": searchTitle})
			return nil
		}
		if isbnMetadata == nil {
			return nil
		}
		if !titleLikelyMatches(title, isbnMetadata.Title) {
			if isbnMetadata.CoverPath != "" {
				_ = os.Remove(isbnMetadata.CoverPath)
			}
			h.Logger.Info("Rejected Calibre external ISBN result for mismatched title", map[string]string{"requested_title": title, "returned_title": isbnMetadata.Title, "isbn": isbn})
			return nil
		}
		if isbnMetadata.ISBN == "" {
			isbnMetadata.ISBN = isbn
		}
		return isbnMetadata
	}

	consider := func(candidate *BookMetadataCandidate) bool {
		if candidate == nil {
			return false
		}
		if candidate.CoverPath != "" {
			metadata = candidate
			return true
		}
		if !metadataMatches {
			metadata = candidate
			metadataMatches = true
		}
		return false
	}

	for _, candidate := range uniqueCalibreRetryCandidates(retryCandidates) {
		titles := retryTitles(title, metadata.Title, candidate.Title)
		for i := 0; i < len(titles); i++ {
			searchTitle := titles[i]
			isbnMetadata := attempt(searchTitle, candidate.ISBN)
			if isbnMetadata == nil {
				continue
			}
			if consider(isbnMetadata) {
				return metadata, nil
			}

			if !containsTitle(titles, isbnMetadata.Title) {
				titles = append(titles, isbnMetadata.Title)
			}
			if candidate.ISBN != "" {
				if consider(attempt(isbnMetadata.Title, "")) {
					return metadata, nil
				}
			}
		}

		if candidate.ISBN != "" && candidate.Title != "" {
			if consider(attempt(candidate.Title, "")) {
				return metadata, nil
			}
		}
	}

	return metadata, nil
}

func (h *Handler) fetchCalibreMetadata(ctx context.Context, coverDir, tempRoot, title, author, isbn string) (*BookMetadataCandidate, error) {
	tempDir, err := os.MkdirTemp(tempRoot, "metadata-*")
	if err != nil {
		return nil, err
	}

	coverPath := filepath.Join(coverDir, calibreCoverFilename(title, author, isbn))
	opfPath := filepath.Join(tempDir, "metadata.opf")
	opfFile, err := os.Create(opfPath)
	if err != nil {
		return nil, err
	}

	args := make([]string, 0, 10)
	if isbn != "" {
		args = append(args, "--isbn", isbn)
	}
	args = append(args, "--title", title)
	if author != "" {
		args = append(args, "--authors", author)
	}
	args = append(args, "--cover", coverPath, "--opf")

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "fetch-ebook-metadata", args...)
	h.Logger.Info("COMMAND: "+cmd.String(), nil)
	cmd.Stdout = opfFile
	cmd.Stderr = &stderr
	cmd.Dir = tempDir

	runErr := cmd.Run()
	closeErr := opfFile.Close()
	if runErr != nil {
		stderrText := strings.TrimSpace(stderr.String())
		if stderrText != "" {
			return nil, fmt.Errorf("calibre metadata fetch failed: %w: %s", runErr, stderrText)
		}
		return nil, fmt.Errorf("calibre metadata fetch failed: %w", runErr)
	}
	if closeErr != nil {
		return nil, closeErr
	}

	metadata, err := h.ParseMetadata(opfPath)
	if err != nil {
		return nil, err
	}
	if metadata.SourceID == "" {
		metadata.SourceID = calibreHash(title, author, isbn)
	}
	if fileExists(coverPath) {
		metadata.CoverPath = coverPath
		h.Logger.Info("Calibre created a cover file", map[string]string{"title": metadata.Title, "isbn": isbn, "cover_path": coverPath})
	} else {
		h.Logger.Info("Calibre did not create a cover file", map[string]string{"title": metadata.Title, "isbn": isbn, "cover_path": coverPath})
	}
	return metadata, nil
}

func (h *Handler) ParseMetadata(opfPath string) (*BookMetadataCandidate, error) {
	file, err := os.Open(opfPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var payload calibreOPFPackage
	if err := xml.NewDecoder(file).Decode(&payload); err != nil {
		return nil, err
	}

	title := strings.TrimSpace(payload.Metadata.Title)
	if title == "" {
		return nil, errors.New("calibre OPF missing title")
	}

	return &BookMetadataCandidate{
		Source:          MetadataSourceCalibre,
		SourceID:        calibreSourceID(payload.Metadata.Identifiers),
		Title:           title,
		Author:          strings.Join(nonEmptyStrings(payload.Metadata.Creators), ", "),
		Description:     cleanCalibreDescription(payload.Metadata.Description),
		Genres:          nonNilStrings(NormalizeGenres(payload.Metadata.Subjects)),
		Language:        firstString(payload.Metadata.Languages),
		PublicationYear: calibrePublicationYear(payload.Metadata.Date),
		ISBN:            calibreISBN(payload.Metadata.Identifiers),
	}, nil
}

func calibreSourceID(identifiers []calibreOPFIdentifier) string {
	preferredSchemes := []string{"ISBN", "GOOGLE", "UUID", "CALIBRE"}

	for _, scheme := range preferredSchemes {
		for _, identifier := range identifiers {
			if strings.EqualFold(identifier.scheme(), scheme) {
				value := strings.TrimSpace(identifier.Value)
				if strings.EqualFold(scheme, "ISBN") {
					value = normalizeISBN(value)
				}
				if value != "" {
					return strings.ToLower(scheme) + ":" + value
				}
			}
		}
	}

	return ""
}

func calibreISBN(identifiers []calibreOPFIdentifier) string {
	for _, identifier := range identifiers {
		if strings.EqualFold(identifier.scheme(), "ISBN") {
			return normalizeISBN(identifier.Value)
		}
	}
	return ""
}

func uniqueISBNs(values []string) []string {
	seen := make(map[string]struct{})
	isbns := make([]string, 0, len(values))
	for _, value := range values {
		isbn := normalizeISBN(value)
		if isbn == "" {
			continue
		}
		if _, ok := seen[isbn]; ok {
			continue
		}
		seen[isbn] = struct{}{}
		isbns = append(isbns, isbn)
	}
	return isbns
}

func uniqueCalibreRetryCandidates(candidates []calibreRetryCandidate) []calibreRetryCandidate {
	seen := make(map[string]struct{})
	unique := make([]calibreRetryCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		candidate.Title = strings.TrimSpace(candidate.Title)
		candidate.ISBN = normalizeISBN(candidate.ISBN)
		if candidate.Title == "" && candidate.ISBN == "" {
			continue
		}
		key := strings.ToLower(candidate.Title) + "\x00" + candidate.ISBN
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, candidate)
	}
	return unique
}

func retryTitles(originalTitle, currentMetadataTitle, candidateTitle string) []string {
	titles := []string{originalTitle, candidateTitle}
	if titleLikelyMatches(originalTitle, currentMetadataTitle) {
		titles = append(titles, currentMetadataTitle)
	}
	return uniqueTitles(titles)
}

func uniqueTitles(values []string) []string {
	seen := make(map[string]struct{})
	unique := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}

func containsTitle(values []string, title string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(title)) {
			return true
		}
	}
	return false
}

func (id calibreOPFIdentifier) scheme() string {
	for _, attr := range id.Attrs {
		if attr.Name.Local == "scheme" {
			return strings.TrimSpace(attr.Value)
		}
	}

	return ""
}

func calibrePublicationYear(value string) *int32 {
	value = strings.TrimSpace(value)
	if len(value) < 4 {
		return nil
	}

	year, err := strconv.ParseInt(value[:4], 10, 32)
	if err != nil || year <= 0 {
		return nil
	}

	publicationYear := int32(year)
	return &publicationYear
}

func cleanCalibreDescription(value string) string {
	value = html.UnescapeString(strings.TrimSpace(value))
	value = htmlTagRE.ReplaceAllString(value, "")
	return strings.TrimSpace(value)
}

func calibreCoverFilename(title, author, isbn string) string {
	base := storage.SanitizeFileName(title)
	if base == "file" {
		base = "cover"
	}

	return fmt.Sprintf("%s_calibre_%s.jpg", base, calibreHash(title, author, isbn))
}

func normalizeISBN(value string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(value) {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == 'x' || r == 'X':
			b.WriteRune('X')
		}
	}
	return b.String()
}

func titleLikelyMatches(queryTitle, candidateTitle string) bool {
	queryTokens := titleTokens(primaryTitle(queryTitle))
	candidateTokens := titleTokens(primaryTitle(candidateTitle))
	if len(queryTokens) == 0 || len(candidateTokens) == 0 {
		return false
	}

	required := queryTokens
	if len(required) > 2 {
		required = required[:2]
	}

	candidateSet := make(map[string]struct{}, len(candidateTokens))
	for _, token := range candidateTokens {
		candidateSet[token] = struct{}{}
	}
	for _, token := range required {
		if _, ok := candidateSet[token]; !ok {
			return false
		}
	}
	return true
}

func primaryTitle(title string) string {
	if before, _, ok := strings.Cut(title, ":"); ok {
		return before
	}
	return title
}

func titleTokens(title string) []string {
	stopWords := map[string]struct{}{
		"a": {}, "an": {}, "and": {}, "the": {},
		"de": {}, "del": {}, "el": {}, "la": {}, "las": {}, "los": {}, "y": {},
		"book": {}, "libro": {}, "saga": {}, "vol": {}, "volumen": {},
	}

	fields := strings.FieldsFunc(strings.ToLower(title), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if _, ok := stopWords[field]; ok {
			continue
		}
		if field != "" {
			tokens = append(tokens, field)
		}
	}
	return tokens
}

func calibreHash(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])[:12]
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Size() > 0
}
