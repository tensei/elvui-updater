package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

var installAddon string

func init() {
	flag.StringVar(&installAddon, "install", "", "elvui and tukui only for now")
	flag.Parse()
}

func main() {
	settings := getSettings()
	clientTukuiAddons := getExtraUIAddons("elvui", "tukui")

	if installAddon != "" {
		install(settings, clientTukuiAddons)
		return
	}

	tukuiAddons := getTukuiAddonList()

	addonsFolder := filepath.Join(settings.WowDirectory, "interface", "addons")
	addons := getLocalTukuiAddons(addonsFolder)
	for _, addon := range addons {

		version := ""
		downloadUrl := ""

		if addon.Toc.XTukuiProjectID < 0 {
			clienttukuiaddon := getClientAddonbyID(clientTukuiAddons, addon.Toc.XTukuiProjectID)
			if clienttukuiaddon != nil {
				version = clienttukuiaddon.Version
				downloadUrl = clienttukuiaddon.URL
			}
		} else {
			tukuiaddon := getAddonbyID(tukuiAddons, fmt.Sprintf("%d", addon.Toc.XTukuiProjectID))
			if tukuiaddon != nil {
				version = tukuiaddon.Version
				downloadUrl = tukuiaddon.URL
			}
		}

		if version == "" || downloadUrl == "" {
			log.Println(addon.Name, "couldn't find the addon in the api")
			continue
		}

		addonVersion, err := strconv.ParseFloat(version, 64)
		if err != nil {
			log.Println(err)
			continue
		}
		if addonVersion > addon.Toc.Version {
			log.Println("updating", addon.Name)
			err = addon.Update(downloadUrl, addonsFolder)
			if err != nil {
				log.Println("error updating", addon.Name, err.Error())
				continue
			}
			updateToc(addon.Toc.path, addon.Toc.XTukuiProjectID)
			log.Printf("finished updating %s from %.2f to %.2f", addon.Name, addon.Toc.Version, addonVersion)
			continue
		}
		log.Println(addon.Name, "is up-to-date!")
	}

	if runtime.GOOS == "windows" {
		fmt.Println("Press the Enter Key to terminate the console screen!")
		var input string
		fmt.Scanln(&input)
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

func getTocFilepath(addonFolder string) string {
	files, _ := readDirIndex(addonFolder)
	for _, file := range files {
		if strings.HasSuffix(file, ".toc") {
			return filepath.Join(addonFolder, file)
		}
	}
	return ""
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

func downloadAddon(link, name string) error {
	log.Println("downloading...", name)
	// Create the file
	out, err := os.Create(name)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := client.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

var addonZipFileName = regexp.MustCompile("[^a-zA-Z0-9]")

func (addon *LocalTukuiAddon) Update(link, addonsPath string) error {
	zipname := fmt.Sprintf("%s.zip", addonZipFileName.ReplaceAllString(addon.Name, ""))
	err := downloadAddon(link, zipname)
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

func install(settings *Settings, addons []ClientApiAddon) {
	addonsFolder := filepath.Join(settings.WowDirectory, "interface", "addons")
	addon := -1
	switch installAddon {
	case "elvui":
		addon = 0
	case "tukui":
		addon = 1
	default:
		log.Println(installAddon, "no supported yet...")
		return
	}
	a := addons[addon]
	zipname := fmt.Sprintf("%s.zip", addonZipFileName.ReplaceAllString(a.Name, ""))
	defer os.Remove(zipname)
	err := downloadAddon(a.URL, zipname)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = Unzip(zipname, addonsFolder)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("done")
}
