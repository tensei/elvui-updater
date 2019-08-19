package main

import (
	"encoding/json"
	"io/ioutil"
)

type Settings struct {
	WowDirectory string `json:"wow_directory"`
}

func getSettings() (*Settings, error) {
	data, err := ioutil.ReadFile("settings.json")
	if err != nil {
		return nil, err
	}
	var settings Settings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}
