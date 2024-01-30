// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"encoding/json"
	"os"
)

func ConfigRead(filename string, instance interface{}) error {
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, instance)
	if err != nil {
		return err
	}

	return nil
}

func ConfigWrite(filename string, instance interface{}) error {
	jsonData, err := json.MarshalIndent(instance, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonData, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
