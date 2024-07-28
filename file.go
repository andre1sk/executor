package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func ParseJsonFileInto(path string, obj interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open JSON file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(obj); err != nil {
		return fmt.Errorf("failed to decode JSON file: %v", err)
	}

	return nil
}
