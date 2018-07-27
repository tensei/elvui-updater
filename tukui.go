package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

var client = http.Client{}

type TukuiAddon struct {
	ID        string `json:"id"`
	Downloads string `json:"downloads"`
	DefaultAddonFields
}

type ClientApiAddon struct {
	ID        int `json:"id"`
	Downloads int `json:"downloads"`
	DefaultAddonFields
}

type DefaultAddonFields struct {
	Name          string `json:"name"`
	Author        string `json:"author"`
	URL           string `json:"url"`
	Version       string `json:"version"`
	Changelog     string `json:"changelog"`
	Ticket        string `json:"ticket"`
	Git           string `json:"git"`
	Patch         string `json:"patch"`
	Lastupdate    string `json:"lastupdate"`
	WebURL        string `json:"web_url"`
	Lastdownload  string `json:"lastdownload"`
	DonateURL     string `json:"donate_url"`
	SmallDesc     string `json:"small_desc"`
	ScreenshotURL string `json:"screenshot_url"`
	Category      string `json:"category"`
}

func getTukuiAddonList() []TukuiAddon {
	api := "https://www.tukui.org/api.php?addons=all"
	req := newRequest("GET", api)
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var addons []TukuiAddon
	err = json.Unmarshal(body, &addons)
	if err != nil {
		log.Fatal(err)
	}
	return addons
}

func newRequest(method, url string) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
	return req
}

func getUIAddon(uis ...string) []ClientApiAddon {
	addons := []ClientApiAddon{}
	for _, ui := range uis {
		api := "https://www.tukui.org/client-api.php?ui=" + ui
		req := newRequest("GET", api)
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		if res.Body != nil {
			defer res.Body.Close()
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}

		var addon ClientApiAddon
		err = json.Unmarshal(body, &addon)
		if err != nil {
			log.Fatal(err)
		}
		addons = append(addons, addon)
	}
	return addons

}
