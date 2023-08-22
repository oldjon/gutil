package csv

import "errors"

var (
	errCsvInvalidPrimaryKeyColumn = errors.New("invalid csv primary key column")
	errCsvInvalidFormat           = errors.New("invalid csv format")
	errCsvInvalidData             = errors.New("invalid csv data")
)
