package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Settings struct {
	WowDirectory string `json:"wow_directory"`
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

func (s *Settings) Save() {
	b, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		log.Println(err)
		return
	}
	ioutil.WriteFile("settings.json", b, 0755)
}
