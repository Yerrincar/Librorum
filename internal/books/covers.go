package books

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"Librorum/internal/storage"
)

type Manager struct {
	CacheDir    string
	HTTPClient  *http.Client
	OpenLibrary OpenLibraryClient
}

func NewManager(cacheDir, openLibraryContact string) *Manager {
	client := &http.Client{Timeout: 10 * time.Second}
	return &Manager{
		CacheDir:   cacheDir,
		HTTPClient: client,
		OpenLibrary: OpenLibraryClient{
			HTTPClient: client,
			Contact:    openLibraryContact,
		},
	}
}

func (m *Manager) Process(ctx context.Context, pkg *Package) (string, error) {
	if pkg == nil {
		return "", fmt.Errorf("metadata package is nil")
	}
	if strings.TrimSpace(m.CacheDir) == "" {
		return "", fmt.Errorf("cover cache directory is empty")
	}
	if err := os.MkdirAll(m.CacheDir, 0o755); err != nil {
		return "", err
	}

	if coverPath := GoodQualityCover(pkg); coverPath != "" {
		cached, err := m.extractFromEPUB(pkg, coverPath)
		if err != nil {
			return "", err
		}
		if cached != "" {
			return cached, nil
		}
	}

	if pkg.InternalCoverPath != "" {
		cached, err := m.extractFromEPUB(pkg, pkg.InternalCoverPath)
		if err != nil {
			return "", err
		}
		if cached != "" {
			return cached, nil
		}
	}

	fallbackPath := m.cachePath(pkg, ".jpg")
	if _, err := os.Stat(fallbackPath); err == nil {
		return fallbackPath, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	coverID, err := m.OpenLibrary.SearchCoverID(ctx, pkg.Metadata.Title, pkg.Metadata.Author)
	if err != nil || coverID == 0 {
		return "", err
	}
	return m.downloadOpenLibraryCover(ctx, coverID, fallbackPath)
}

func GoodQualityCover(pkg *Package) string {
	const dimensionCap = 2.0 / 3.0
	const minUniqueColors = 5

	ignoredNameTokens := []string{
		"title", "endpaper", "endpapers", "backad", "adcard", "abouttheauthor",
		"newsletter", "contents", "toc", "copyright", "frontmatter", "backmatter",
	}

	z, err := zip.OpenReader(pkg.SourcePath)
	if err != nil {
		return ""
	}
	defer z.Close()

	possibleCovers := make([]*zip.File, 0)
	for _, file := range z.File {
		ext := strings.ToLower(path.Ext(file.Name))
		if ext != ".jpg" && ext != ".jpeg" {
			continue
		}

		lowerName := strings.ToLower(file.Name)
		if containsAny(lowerName, ignoredNameTokens) {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			continue
		}
		cfg, err := jpeg.DecodeConfig(rc)
		rc.Close()
		if err != nil {
			continue
		}

		switch {
		case cfg.Width <= 400 || cfg.Height <= 600:
			continue
		case float64(cfg.Width)/float64(cfg.Height) > dimensionCap:
			continue
		case cfg.Width == cfg.Height:
			continue
		default:
			possibleCovers = append(possibleCovers, file)
		}
	}

	winner := ""
	bestScore := 0
	for _, candidate := range possibleCovers {
		score := scoreCoverCandidate(candidate, pkg.InternalCoverPath, dimensionCap, minUniqueColors)
		if score > bestScore {
			bestScore = score
			winner = candidate.Name
		}
	}
	return winner
}

func scoreCoverCandidate(candidate *zip.File, internalCoverPath string, dimensionCap float64, minUniqueColors int) int {
	rc, err := candidate.Open()
	if err != nil {
		return 0
	}
	defer rc.Close()

	img, err := jpeg.Decode(rc)
	if err != nil {
		return 0
	}

	score := img.Bounds().Dx() * img.Bounds().Dy()
	if internalCoverPath != "" && strings.EqualFold(path.Clean(candidate.Name), path.Clean(internalCoverPath)) {
		score *= 4
	}
	if strings.Contains(strings.ToLower(candidate.Name), "cover") {
		score *= 2
	}
	if float64(img.Bounds().Dx())/float64(img.Bounds().Dy()) == dimensionCap {
		score *= 3
	}

	colors := make(map[color.Color]int)
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x += 50 {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y += 50 {
			colors[img.At(x, y)]++
		}
	}
	if len(colors) >= minUniqueColors {
		score += len(colors)
	}
	return score
}

func (m *Manager) extractFromEPUB(pkg *Package, internalPath string) (string, error) {
	z, err := zip.OpenReader(pkg.SourcePath)
	if err != nil {
		return "", err
	}
	defer z.Close()

	for _, file := range z.File {
		if !strings.EqualFold(path.Clean(file.Name), path.Clean(internalPath)) {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		cachedPath := m.cachePath(pkg, coverExtension(internalPath))
		out, err := os.Create(cachedPath)
		if err != nil {
			return "", err
		}
		defer out.Close()

		if _, err := io.Copy(out, rc); err != nil {
			return "", err
		}
		return cachedPath, out.Sync()
	}

	return "", nil
}

func (m *Manager) downloadOpenLibraryCover(ctx context.Context, coverID int, dstPath string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://covers.openlibrary.org/b/id/%d.jpg", coverID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", m.OpenLibrary.UserAgent())

	client := m.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openlibrary cover download failed: %s", resp.Status)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	return dstPath, out.Sync()
}

func (m *Manager) cachePath(pkg *Package, ext string) string {
	if ext == "" {
		ext = ".jpg"
	}
	title := storage.SanitizeFileName(pkg.Metadata.Title)
	if title == "file" {
		title = storage.SanitizeFileName(strings.TrimSuffix(pkg.FileName, filepath.Ext(pkg.FileName)))
	}
	sum := sha256.Sum256([]byte(pkg.SourcePath))
	return filepath.Join(m.CacheDir, fmt.Sprintf("%s_%s%s", title, hex.EncodeToString(sum[:])[:12], ext))
}

func coverExtension(internalPath string) string {
	ext := strings.ToLower(path.Ext(internalPath))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return ext
	default:
		return ".jpg"
	}
}

func containsAny(value string, tokens []string) bool {
	for _, token := range tokens {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}
