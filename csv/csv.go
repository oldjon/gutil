package csv

import (
	"fmt"
	"github.com/oldjon/gutil"
	"io/ioutil"
	"strings"
)

// ReadCSV read a csv file to [][]string
func ReadCSV(fileName string) ([][]string, error) {
	cf, err := NewCSVFileFromFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open csv file, %w", err)
	}

	return cf.rawRecords, nil
}

// ReadCSV2MapWithPrimaryKey read a csv file to map[string]map[string]string
func ReadCSV2MapWithPrimaryKey(fileName string, primaryKeyColumn int) (map[string]map[string]string, error) {
	cf, err := NewCSVFileFromFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open csv file, %w", err)
	}

	m, err := cf.AsMapWithPrimaryKey(primaryKeyColumn)
	if err != nil {
		return nil, fmt.Errorf("failed to call AsMapWithPrimaryKey, %w", err)
	}

	return m, nil
}

// ReadCSV2Array read a csv file to []map[string]string
func ReadCSV2Array(fileName string) ([]map[string]string, error) {
	cf, err := NewCSVFileFromFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open csv file, %w", err)
	}

	a, err := cf.AsArray()
	if err != nil {
		return nil, fmt.Errorf("failed to call AsArray, %w", err)
	}

	return a, nil
}

// GetAllCsvFiles get all csv files from current dir not sub dirs
func GetAllCsvFiles(dirPath string) ([]string, error) {
	fileList := make([]string, 0)
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, fi := range files {
		if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".csv") {
			fileList = append(fileList, fi.Name())
		}

	}
	return fileList, nil
}

// CheckCSVPrefix check file whether matched prefix. return false when not matched. return the remaining part(exclude ".csv" suffix) and true when matched.
func CheckCSVPrefix(file string, prefix string) (string, bool) {
	if strings.Index(file, prefix) != 0 {
		return "", false
	}
	beginIndex := len(prefix)
	endIndex := strings.Index(file, ".csv")
	endIndex = gutil.If(endIndex == -1, len(file)-1, endIndex)
	midPart := strings.ToUpper(file[beginIndex : endIndex+1])
	return midPart, true
}
