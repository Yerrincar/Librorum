package metadata

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Package struct {
	Metadata          Metadata `xml:"metadata"`
	Manifest          Manifest `xml:"manifest"`
	Guide             Guide    `xml:"guide"`
	InternalCoverPath string
	SourcePath        string
	FileName          string
}

type Metadata struct {
	Author      string   `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Title       string   `xml:"http://purl.org/dc/elements/1.1/ title"`
	Description string   `xml:"http://purl.org/dc/elements/1.1/ description"`
	Genres      []string `xml:"http://purl.org/dc/elements/1.1/ subject"`
	Language    string   `xml:"http://purl.org/dc/elements/1.1/ language"`
	Metas       []Meta   `xml:"meta"`
}

type Meta struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

type Manifest struct {
	Items []Item `xml:"item"`
}

type Item struct {
	ID         string `xml:"id,attr"`
	Href       string `xml:"href,attr"`
	Properties string `xml:"properties,attr"`
}

type Guide struct {
	References []Reference `xml:"reference"`
}

type Reference struct {
	Type  string `xml:"type,attr"`
	Href  string `xml:"href,attr"`
	Title string `xml:"title,attr"`
}

func ExtractEPUB(srcPath string) (*Package, error) {
	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return nil, fmt.Errorf("open epub: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if !strings.HasSuffix(strings.ToLower(f.Name), ".opf") {
			continue
		}

		pkg, err := extractOPF(r, f)
		if err != nil {
			return nil, err
		}
		pkg.SourcePath = srcPath
		pkg.FileName = filepath.Base(srcPath)
		pkg.Metadata.Genres = NormalizeGenres(pkg.Metadata.Genres)
		return pkg, nil
	}

	return nil, fmt.Errorf("epub metadata .opf file not found: %s", srcPath)
}

func NormalizeGenres(genres []string) []string {
	if len(genres) == 0 {
		return nil
	}

	out := make([]string, 0, len(genres))
	seen := make(map[string]struct{}, len(genres))
	for _, genre := range genres {
		genre = strings.TrimSpace(strings.Trim(genre, ","))
		if genre == "" {
			continue
		}
		key := strings.ToLower(genre)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, genre)
	}
	return out
}

func extractOPF(r *zip.ReadCloser, f *zip.File) (*Package, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("open opf: %w", err)
	}
	defer rc.Close()

	raw, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read opf: %w", err)
	}

	var pkg Package
	if err := xml.Unmarshal(raw, &pkg); err != nil {
		return nil, fmt.Errorf("parse opf xml: %w", err)
	}

	baseDir := path.Dir(f.Name)
	coverID := ""
	coverGuideHref := ""
	for _, meta := range pkg.Metadata.Metas {
		if meta.Name == "cover" && meta.Content != "" {
			coverID = meta.Content
			break
		}
	}
	for _, ref := range pkg.Guide.References {
		if ref.Type == "cover" && ref.Href != "" {
			coverGuideHref = ref.Href
			break
		}
	}

	for _, item := range pkg.Manifest.Items {
		if (coverGuideHref != "" && item.Href == coverGuideHref) ||
			(coverID != "" && item.ID == coverID) ||
			strings.Contains(item.Properties, "cover-image") ||
			item.ID == "cover" {
			pkg.InternalCoverPath = path.Join(baseDir, item.Href)
			break
		}
	}
	if pkg.InternalCoverPath == "" && coverGuideHref != "" {
		pkg.InternalCoverPath = path.Join(baseDir, coverGuideHref)
	}
	if pkg.InternalCoverPath != "" {
		resolved, err := resolveCoverPath(r, pkg.InternalCoverPath)
		if err == nil && resolved != "" {
			pkg.InternalCoverPath = resolved
		}
	}

	return &pkg, nil
}

func resolveCoverPath(r *zip.ReadCloser, href string) (string, error) {
	ext := strings.ToLower(path.Ext(href))
	if ext != ".xhtml" && ext != ".html" && ext != ".xml" {
		return href, nil
	}

	imgRel, err := resolveCoverFromXHTML(r, href)
	if err != nil || imgRel == "" {
		return "", err
	}
	return path.Join(path.Dir(href), imgRel), nil
}

func resolveCoverFromXHTML(r *zip.ReadCloser, href string) (string, error) {
	f, err := findZipFile(r, href)
	if err != nil {
		return "", err
	}

	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	dec := xml.NewDecoder(rc)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || !strings.EqualFold(se.Name.Local, "img") {
			continue
		}
		for _, attr := range se.Attr {
			if strings.EqualFold(attr.Name.Local, "src") && attr.Value != "" {
				return attr.Value, nil
			}
		}
	}
}

func findZipFile(r *zip.ReadCloser, name string) (*zip.File, error) {
	cleanName := path.Clean(name)
	for _, f := range r.File {
		if path.Clean(f.Name) == cleanName {
			return f, nil
		}
	}
	return nil, os.ErrNotExist
}
