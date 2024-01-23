// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"encoding/json"
	"os"
)

func ConfigRead(filename string) error {
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &Options)
	if err != nil {
		return err
	}

	return nil
}

func ConfigWrite(filename string) error {
	jsonData, err := json.MarshalIndent(&Options, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonData, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
