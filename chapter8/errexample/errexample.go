package errexample

import (
	"errors"
	"fmt"
	"log"
	"os"
)

var customErr = errors.New("Non existing file, can not open")

func FileOpenRead(fpath string) {
	// Read the file
	data, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	content := string(data)
	fmt.Println(content)
}

func FileOpenRead2(fpath string) (string, error) {
	// Read the file
	data, err := os.ReadFile(fpath)
	if err != nil {
		return "", customErr
	}

	content := string(data)
	return content, nil
}
