package books

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ExcelBookRow struct {
	Title  string
	Author string
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

	titleRow := -1
	titleCol := -1
	authorCol := -1
	rows, err := f.GetRows(spreadsheet)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for i, row := range rows {
		for j, col := range row {
			header := strings.ToLower(strings.TrimSpace(col))
			if header == "title" {
				titleRow = i
				titleCol = j
			}
			if header == "author" {
				authorCol = j
			}
		}
		if titleCol >= 0 && authorCol >= 0 {
			break
		}
	}
	if titleRow < 0 || titleCol < 0 || authorCol < 0 {
		return nil, errors.New("spreadsheet must contain title and author columns")
	}

	titleAuthorRows := make([]ExcelBookRow, 0, len(rows)-titleRow-1)
	for i := titleRow + 1; i < len(rows); i++ {
		if len(rows[i]) <= titleCol {
			continue
		}
		titleValue := strings.TrimSpace(rows[i][titleCol])
		if titleValue == "" {
			continue
		}
		authorValue := ""
		if len(rows[i]) > authorCol {
			authorValue = strings.TrimSpace(rows[i][authorCol])
		}
		titleAuthorRows = append(titleAuthorRows, ExcelBookRow{Title: titleValue, Author: authorValue})
	}
	return titleAuthorRows, nil
}
