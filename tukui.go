package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"sync"
)

var client = http.Client{}

// TukuiAddon ...
type TukuiAddon struct {
	ID        string `json:"id"`
	Downloads string `json:"downloads"`
	DefaultAddonFields
}

// ClientAPIAddon ...
type ClientAPIAddon struct {
	ID        int `json:"id"`
	Downloads int `json:"downloads"`
	DefaultAddonFields
}

// DefaultAddonFields ...
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

// LocalTukuiAddon ...
type LocalTukuiAddon struct {
	Name string
	Path string
	Toc  *Toc
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

func getExtraUIAddons(uis ...string) []ClientAPIAddon {
	addons := []ClientAPIAddon{}
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

		var addon ClientAPIAddon
		err = json.Unmarshal(body, &addon)
		if err != nil {
			log.Fatal(err)
		}
		addons = append(addons, addon)
	}
	return addons

}

func getLocalTukuiAddons(addonsPath string) []LocalTukuiAddon {
	localAddons := []LocalTukuiAddon{}
	addons, err := readDirIndex(addonsPath)
	if err != nil {
		log.Println(err)
		return localAddons
	}
	addonChan := make(chan string, len(addons))
	wg := sync.WaitGroup{}
	wg.Add(len(addons))

	max := runtime.NumCPU()
	for i := 0; i < max; i++ {
		go func() {
			for addon := range addonChan {
				addonPath := filepath.Join(addonsPath, addon)
				tocfile := getTocFilepath(addonPath)
				if tocfile == "" {
					wg.Done()
					continue
				}
				toc := parseToc(tocfile)
				if !toc.HasProjectID {
					wg.Done()
					continue
				}
				localAddons = append(localAddons, LocalTukuiAddon{
					Name: addon,
					Path: addonPath,
					Toc:  toc,
				})
				wg.Done()
			}
		}()
	}
	for _, addon := range addons {
		addonChan <- addon
	}
	wg.Wait()
	return localAddons
}
