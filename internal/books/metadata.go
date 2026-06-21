package books

const (
	MetadataSourceOpenLibrary = "openlibrary"
	MetadataSourceGoogleBooks = "google_books"
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
	CoverID         int      `json:"cover_id"`
	CoverURL        string   `json:"cover_url"`
	WorkKey         string   `json:"work_key"`
}
