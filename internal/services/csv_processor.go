package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

type CSVProcessor struct{}

func NewCSVProcessor() *CSVProcessor {
	return &CSVProcessor{}
}

func (p *CSVProcessor) Process(r io.Reader) ([]interface{}, error) {
	reader := csv.NewReader(r)

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	headers = normalizeHeaders(headers)

	var records []interface{}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		record := make(map[string]string)
		for i, header := range headers {
			if i < len(row) {
				record[header] = strings.TrimSpace(row[i])
			}
		}

		records = append(records, record)
	}

	return records, nil
}

func normalizeHeaders(headers []string) []string {
	normalized := make([]string, len(headers))
	for i, h := range headers {
		normalized[i] = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(h, " ", "_")))
	}
	return normalized
}

func (p *CSVProcessor) ValidateHeaders(headers []string, requiredFields []string) error {
	headerMap := make(map[string]bool)
	for _, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = true
	}

	for _, field := range requiredFields {
		if !headerMap[field] {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	return nil
}
