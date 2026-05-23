package kindle

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"Librorum/internal/storage"
)

var (
	ErrKindleNotFound = errors.New("kindle mtp mount not found")
	mtpRootRe         = regexp.MustCompile(`default_location=(mtp://Amazon_Kindle_[^/]+/)`)
)

type SyncResult struct {
	DetectedBooks []string
	ImportedFiles []string
	Failed        int
	Duplicated    int
}

func DetectRootURI(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gio", "mount", "-li")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gio mount -li failed: %w\n%s", err, string(out))
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		match := mtpRootRe.FindStringSubmatch(line)
		if len(match) == 2 {
			return match[1], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", ErrKindleNotFound
}

func DocumentsURI(root string) string {
	return strings.TrimRight(root, "/") + "/Internal Storage/documents/"
}

func ListFiles(ctx context.Context, docsURI string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "gio", "list", docsURI)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gio list failed: %w\n%s", err, string(out))
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" || strings.HasSuffix(name, "/") {
			continue
		}
		files = append(files, name)
	}
	return files, nil
}

func FilterConvertible(entries []string) []string {
	allowed := map[string]struct{}{
		".epub": {},
		".azw":  {},
		".azw3": {},
		".mobi": {},
		".pdf":  {},
		".txt":  {},
	}

	filtered := make([]string, 0, len(entries))
	for _, entry := range entries {
		if _, ok := allowed[strings.ToLower(filepath.Ext(entry))]; ok {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func ScanBooks(ctx context.Context) (docsURI string, books []string, err error) {
	root, err := DetectRootURI(ctx)
	if err != nil {
		return "", nil, err
	}
	docsURI = DocumentsURI(root)
	entries, err := ListFiles(ctx, docsURI)
	if err != nil {
		return "", nil, err
	}
	return docsURI, FilterConvertible(entries), nil
}

func CopySelected(ctx context.Context, docsURI string, selected []string, destinationDir string) (SyncResult, error) {
	var result SyncResult
	if strings.TrimSpace(destinationDir) == "" {
		return result, fmt.Errorf("destination directory is empty")
	}
	if docsURI == "" {
		root, err := DetectRootURI(ctx)
		if err != nil {
			return result, err
		}
		docsURI = DocumentsURI(root)
	}

	detected, err := ListFiles(ctx, docsURI)
	if err != nil {
		return result, err
	}
	detected = FilterConvertible(detected)
	result.DetectedBooks = detected

	target := selected
	if len(target) == 0 {
		target = detected
	}

	existing, err := existingFileNames(destinationDir)
	if err != nil {
		return result, err
	}

	tmpDir, err := os.MkdirTemp("", "librorum-kindle-sync-*")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(tmpDir)

	for _, name := range target {
		outputName := outputEPUBName(name)
		if _, ok := existing[outputName]; ok {
			result.Duplicated++
			continue
		}

		localSrc := filepath.Join(tmpDir, storage.SanitizeFileName(name))
		if err := CopyFromKindle(ctx, JoinMTP(docsURI, name), localSrc); err != nil {
			result.Failed++
			continue
		}

		finalSrc := localSrc
		if strings.ToLower(filepath.Ext(name)) != ".epub" {
			converted := filepath.Join(tmpDir, outputName)
			if err := convertToEPUB(ctx, localSrc, converted); err != nil {
				result.Failed++
				continue
			}
			finalSrc = converted
		}

		dstPath := filepath.Join(destinationDir, outputName)
		if err := storage.CopyFile(finalSrc, dstPath); err != nil {
			result.Failed++
			continue
		}

		existing[outputName] = struct{}{}
		result.ImportedFiles = append(result.ImportedFiles, dstPath)
	}

	return result, nil
}

func CopyFromKindle(ctx context.Context, srcURI, localDstPath string) error {
	absDst, err := filepath.Abs(localDstPath)
	if err != nil {
		return err
	}
	dstURI := (&url.URL{Scheme: "file", Path: absDst}).String()

	cmd := exec.CommandContext(ctx, "gio", "copy", "--", srcURI, dstURI)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gio copy failed: %w\n%s", err, string(out))
	}
	return nil
}

func JoinMTP(baseURI, fileName string) string {
	return strings.TrimRight(baseURI, "/") + "/" + url.PathEscape(fileName)
}

func existingFileNames(dir string) (map[string]struct{}, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	existing := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		existing[entry.Name()] = struct{}{}
	}
	return existing, nil
}

func outputEPUBName(name string) string {
	clean := storage.SanitizeFileName(name)
	if strings.ToLower(filepath.Ext(clean)) == ".epub" {
		return clean
	}
	return strings.TrimSuffix(clean, filepath.Ext(clean)) + ".epub"
}

func convertToEPUB(ctx context.Context, src, dst string) error {
	cmd := exec.CommandContext(ctx, "ebook-convert", src, dst)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ebook-convert failed: %w\n%s", err, string(out))
	}
	return nil
}
