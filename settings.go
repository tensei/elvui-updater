package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Settings struct {
	WowDirectory string   `json:"wow_directory"`
	Addons       []string `json:"addons"`
}

func getSettings() *Settings {
	data, err := ioutil.ReadFile("settings.json")
	if err != nil {
		log.Println(err)
		return nil
	}
	var settings Settings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &settings
}
