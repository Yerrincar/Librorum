package books

const (
	MetadataSourceOpenLibrary = "openlibrary"
	MetadataSourceGoogleBooks = "google_books"
	MetadataSourceCalibre     = "calibre"
)

type BookMetadataCandidate struct {
	Source          string   `json:"source"`
	SourceID        string   `json:"source_id"`
	Title           string   `json:"title"`
	Author          string   `json:"author"`
	Description     string   `json:"description"`
	Genres          []string `json:"genres"`
	Language        string   `json:"language"`
	PublicationYear *int32   `json:"publication_year"`
	ISBN            string   `json:"isbn"`
	ISBNs           []string `json:"isbns,omitempty"`
	CoverID         int      `json:"cover_id"`
	CoverURL        string   `json:"cover_url"`
	CoverPath       string   `json:"cover_path"`
	WorkKey         string   `json:"work_key"`
}
