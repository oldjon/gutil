package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

type CSVFile interface {
	// AsMapWithPrimaryKey convert content to map[string]map[string]string
	AsMapWithPrimaryKey(primaryKeyColumn int) (map[string]map[string]string, error)

	// AsArray convert content to []map[string]string
	AsArray() ([]map[string]string, error)
}

var (
	_ CSVFile = (*csvFile)(nil)
)

// NewCSVFile create a CSVFile from io.Reader
//
//revive:disable:unexported-return
func NewCSVFile(r io.Reader) (*csvFile, error) {
	reader := csv.NewReader(r)
	rawRecords, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read conttent, %w", err)
	}

	return &csvFile{rawRecords: rawRecords}, nil
}

//revive:enable:unexported-return

// NewCSVFileFromFile create a CSVFile from a physic file path
//
//revive:disable:unexported-return
func NewCSVFileFromFile(fileName string) (*csvFile, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file[%s], %w", fileName, err)
	}
	defer file.Close()

	SkipBom(file)

	return NewCSVFile(file)
}

//revive:enable:unexported-return

type csvFile struct {
	rawRecords [][]string
}

// AsMapWithPrimaryKey convert content to map[string]map[string]string
func (c *csvFile) AsMapWithPrimaryKey(primaryKeyColumn int) (map[string]map[string]string, error) {
	rawRecords := c.rawRecords

	return convertCSVRecords2MapWithPrimaryKey(rawRecords, primaryKeyColumn)

}

func convertCSVRecords2MapWithPrimaryKey(rawRecords [][]string, primaryKeyColumn int) (map[string]map[string]string, error) {
	headerRowCount := 3
	if primaryKeyColumn < 0 {
		return nil, errCsvInvalidPrimaryKeyColumn
	}

	if len(rawRecords) < headerRowCount {
		return nil, errCsvInvalidFormat
	}

	headers := rawRecords[0]
	for i, v := range headers {
		headers[i] = strings.ToLower(v)
	}

	if primaryKeyColumn >= len(headers) {
		return nil, errCsvInvalidFormat
	}

	mapRecords := make(map[string]map[string]string)
	for i := headerRowCount; i < len(rawRecords); i++ {
		rawRecord := rawRecords[i]
		if len(rawRecord) < len(headers) {
			return nil, errCsvInvalidData
		}
		mapRecord := make(map[string]string)
		for j, h := range headers {
			mapRecord[h] = rawRecord[j]
		}
		mapRecords[rawRecord[primaryKeyColumn]] = mapRecord
	}
	return mapRecords, nil
}

// AsArray convert content to []map[string]string
func (c *csvFile) AsArray() ([]map[string]string, error) {
	rawRecords := c.rawRecords

	return convertCSVRecords2Array(rawRecords)
}

func convertCSVRecords2Array(rawRecords [][]string) ([]map[string]string, error) {
	headerRowCount := 3

	if len(rawRecords) < headerRowCount {
		return nil, errCsvInvalidFormat
	}

	headers := rawRecords[0]
	for i, v := range headers {
		headers[i] = strings.ToLower(v)
	}

	arrayRecords := make([]map[string]string, 0)
	for i := headerRowCount; i < len(rawRecords); i++ {
		rawRecord := rawRecords[i]
		if len(rawRecord) < len(headers) {
			return nil, errCsvInvalidData
		}
		mapRecord := make(map[string]string)
		for j, h := range headers {
			mapRecord[h] = rawRecord[j]
		}
		arrayRecords = append(arrayRecords, mapRecord)
	}
	return arrayRecords, nil
}
