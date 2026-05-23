package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type Paths struct {
	RootDir       string
	BooksDir      string
	CoverCacheDir string
	ImportsDir    string
}

func NewPaths(rootDir string) Paths {
	if strings.TrimSpace(rootDir) == "" {
		rootDir = "data"
	}

	return Paths{
		RootDir:       rootDir,
		BooksDir:      filepath.Join(rootDir, "books"),
		CoverCacheDir: filepath.Join(rootDir, "covers"),
		ImportsDir:    filepath.Join(rootDir, "imports"),
	}
}

func EnsureDirs(paths Paths) error {
	for _, dir := range []string{paths.RootDir, paths.BooksDir, paths.CoverCacheDir, paths.ImportsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func CopyFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return dst.Sync()
}

func FileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func SanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		case r == '.', r == '-', r == '_':
			return r
		case unicode.IsSpace(r):
			return '_'
		default:
			return '_'
		}
	}, name)
	name = strings.Trim(name, "._-")
	if name == "" {
		return "file"
	}
	return name
}

func UniquePath(dir, desiredName string) (string, string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}

	cleanName := SanitizeFileName(filepath.Base(desiredName))
	ext := filepath.Ext(cleanName)
	base := strings.TrimSuffix(cleanName, ext)

	for i := 0; ; i++ {
		candidateName := cleanName
		if i > 0 {
			candidateName = fmt.Sprintf("%s_%d%s", base, i, ext)
		}
		candidatePath := filepath.Join(dir, candidateName)
		_, err := os.Stat(candidatePath)
		if os.IsNotExist(err) {
			return candidatePath, candidateName, nil
		}
		if err != nil {
			return "", "", err
		}
	}
}
