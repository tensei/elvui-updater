package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func main() {
	settings := getSettings()
	tukuiAddons := getTukuiAddonList()
	clientTukuiAddons := getUIAddon("elvui", "tukui")
	addonsFolder := filepath.Join(settings.WowDirectory, "interface", "addons")
	addons, err := readDirIndex(addonsFolder)
	if err != nil {
		log.Fatal(err)
	}
	for _, addon := range addons {
		if _, ok := inSlice(settings.Addons, addon); !ok {
			continue
		}
		addonFolder := filepath.Join(addonsFolder, addon)
		files, err := readDirIndex(addonFolder)
		if err != nil {
			continue
		}
		for _, file := range files {
			if !strings.HasSuffix(file, ".toc") {
				continue
			}
			data, err := ioutil.ReadFile(filepath.Join(addonFolder, file))
			if err != nil {
				log.Println(err)
				break
			}
			toc := parseToc(string(data))

			version := ""
			downloadUrl := ""

			if toc.XTukuiProjectID < 0 {
				clienttukuiaddon := getClientAddonbyID(clientTukuiAddons, toc.XTukuiProjectID)
				if clienttukuiaddon != nil {
					version = clienttukuiaddon.Version
					downloadUrl = clienttukuiaddon.URL
				}
			} else {
				tukuiaddon := getAddonbyID(tukuiAddons, fmt.Sprintf("%d", toc.XTukuiProjectID))
				if tukuiaddon != nil {
					version = tukuiaddon.Version
					downloadUrl = tukuiaddon.URL
				}
			}

			if version == "" || downloadUrl == "" {
				log.Println(addon, "couldn't find the addon in the api")
				break
			}

			addonVersion, err := strconv.ParseFloat(version, 64)
			if err != nil {
				log.Println(err)
				break
			}
			if addonVersion > toc.Version {
				log.Println("updating", addon)
				err = updateAddon(downloadUrl, addon, addonsFolder)
				if err != nil {
					log.Println("error updating", addon, err.Error())
					break
				}
				log.Println("finished updating", addon)
				break
			}
			log.Println(addon, "is up-to-date!")
			break
		}
	}

}

func readDirIndex(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

func inSlice(slice []string, item string) (int, bool) {
	for pos, i := range slice {
		if strings.EqualFold(i, item) {
			return pos, true
		}
	}
	return 0, false
}

func getAddonbyID(addons []TukuiAddon, id string) *TukuiAddon {
	for _, a := range addons {
		if a.ID == id {
			return &a
		}
	}
	return nil
}

func getClientAddonbyID(addons []ClientApiAddon, id int) *ClientApiAddon {
	for _, a := range addons {
		if a.ID == id {
			return &a
		}
	}
	return nil
}

func doanloadAddon(link, name string) error {
	log.Println("downloading...", name)
	req := newRequest("GET", link)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(name, data, 0755)
}

var addonZipFileName = regexp.MustCompile("[^a-zA-Z0-9]")

func updateAddon(link, name, addonsPath string) error {
	zipname := fmt.Sprintf("%s.zip", addonZipFileName.ReplaceAllString(name, ""))
	err := doanloadAddon(link, zipname)
	if err != nil {
		return err
	}
	_, err = Unzip(zipname, addonsPath)
	if err != nil {
		return err
	}
	os.Remove(zipname)
	return nil
}
