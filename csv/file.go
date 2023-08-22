package csv

import (
	"os"
	"path/filepath"
)

// Get compiled executable file absolute path
func SelfPath() string {
	path, _ := filepath.Abs(os.Args[0])
	return path
}

// Get compiled executable file directory
func SelfDir() string {
	return filepath.Dir(SelfPath())
}

// Check if file exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
