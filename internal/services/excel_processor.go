package services

import (
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ExcelProcessor struct{}

func NewExcelProcessor() *ExcelProcessor {
	return &ExcelProcessor{}
}

func (p *ExcelProcessor) Process(r io.Reader) ([]interface{}, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("no sheets found in excel file")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("excel file is empty")
	}

	headers := normalizeExcelHeaders(rows[0])

	var records []interface{}

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) == 0 || isAllEmpty(row) {
			continue
		}

		record := make(map[string]string)
		for j, header := range headers {
			if j < len(row) {
				record[header] = row[j]
			}
		}

		records = append(records, record)
	}

	return records, nil
}

func normalizeExcelHeaders(headers []string) []string {
	normalized := make([]string, len(headers))
	for i, h := range headers {
		normalized[i] = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(h, " ", "_")))
	}
	return normalized
}

func isAllEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
