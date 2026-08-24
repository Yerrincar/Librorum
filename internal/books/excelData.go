package books

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ExcelBookRow struct {
	Title             string
	Author            string
	Kind              string
	OwnershipStatus   string
	ReadingStatus     string
	PublicationStatus string
	ReadAt            string
	Rating            string
	Notes             string
	PublicationYear   string
	Format            string
	CurrentChapter    string
	TotalChapters     string
}

func ExcelTitleAuthor(file, spreadsheet string) ([]ExcelBookRow, error) {
	f, err := excelize.OpenFile(file)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	headerRow := -1
	columns := map[string]int{}
	rows, err := f.GetRows(spreadsheet)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for i, row := range rows {
		rowColumns := map[string]int{}
		for j, col := range row {
			if header, ok := excelHeaderName(col); ok {
				rowColumns[header] = j
			}
		}
		if _, hasTitle := rowColumns["title"]; hasTitle {
			if _, hasAuthor := rowColumns["author"]; hasAuthor {
				headerRow = i
				columns = rowColumns
				break
			}
		}
	}
	if headerRow < 0 {
		return nil, errors.New("spreadsheet must contain title and author columns")
	}

	bookRows := make([]ExcelBookRow, 0, len(rows)-headerRow-1)
	for i := headerRow + 1; i < len(rows); i++ {
		titleValue := excelCell(rows[i], columns, "title")
		if titleValue == "" {
			continue
		}
		bookRows = append(bookRows, ExcelBookRow{
			Title:             titleValue,
			Author:            excelCell(rows[i], columns, "author"),
			Kind:              excelCell(rows[i], columns, "kind"),
			OwnershipStatus:   excelCell(rows[i], columns, "ownership_status"),
			ReadingStatus:     excelCell(rows[i], columns, "reading_status"),
			PublicationStatus: excelCell(rows[i], columns, "publication_status"),
			ReadAt:            excelCell(rows[i], columns, "read_at"),
			Rating:            excelCell(rows[i], columns, "rating"),
			Notes:             excelCell(rows[i], columns, "notes"),
			PublicationYear:   excelCell(rows[i], columns, "publication_year"),
			Format:            excelCell(rows[i], columns, "format"),
			CurrentChapter:    excelCell(rows[i], columns, "current_chapter"),
			TotalChapters:     excelCell(rows[i], columns, "total_chapters"),
		})
	}
	return bookRows, nil
}

func excelCell(row []string, columns map[string]int, name string) string {
	col, ok := columns[name]
	if !ok || len(row) <= col {
		return ""
	}
	return strings.TrimSpace(row[col])
}

func excelHeaderName(value string) (string, bool) {
	header := normalizeExcelHeader(value)
	aliases := map[string]string{
		"title":                           "title",
		"titulo":                          "title",
		"book_title":                      "title",
		"author":                          "author",
		"autor":                           "author",
		"kind":                            "kind",
		"type":                            "kind",
		"tipo":                            "kind",
		"ownership":                       "ownership_status",
		"ownership_status":                "ownership_status",
		"ownership_status_":               "ownership_status",
		"reading_status":                  "reading_status",
		"reading":                         "reading_status",
		"publication_status":              "publication_status",
		"read_at":                         "read_at",
		"read_date":                       "read_at",
		"finished_at":                     "read_at",
		"rating":                          "rating",
		"score":                           "rating",
		"nota":                            "rating",
		"notes":                           "notes",
		"note":                            "notes",
		"notas":                           "notes",
		"publication_year":                "publication_year",
		"publication_date":                "publication_year",
		"published_year":                  "publication_year",
		"year":                            "publication_year",
		"ano":                             "publication_year",
		"format":                          "format",
		"formato":                         "format",
		"current_chapter":                 "current_chapter",
		"current_chapters":                "current_chapter",
		"chapter":                         "current_chapter",
		"total_chapters":                  "total_chapters",
		"total_chapter":                   "total_chapters",
		"ownership_status_reading_status": "ownership_status",
	}
	canonical, ok := aliases[header]
	return canonical, ok
}

func normalizeExcelHeader(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n",
		"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ñ", "n",
		"/", "_", "-", "_", " ", "_", ".", "_",
	)
	value = replacer.Replace(value)
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	return strings.Trim(value, "_")
}
