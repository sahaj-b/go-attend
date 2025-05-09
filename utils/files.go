package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func EnsureAndGetFile(path string, mode string) (dataFile *os.File, err error) {
	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}
	flags := os.O_CREATE
	switch mode {
	case "r":
		flags |= os.O_RDONLY
	case "w":
		flags |= os.O_WRONLY
	case "a":
		flags |= os.O_APPEND | os.O_WRONLY
	case "rw":
		flags |= os.O_RDWR
	default:
		return nil, fmt.Errorf("Invalid mode: %v", mode)
	}

	dataFile, err = os.OpenFile(path, flags, 0644)
	if err != nil {
		return nil, err
	}
	return dataFile, nil
}
