package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

var installAddon string

func init() {
	flag.StringVar(&installAddon, "install", "", "elvui and tukui only for now")
	flag.Parse()
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
}

func main() {
	settings, err := getSettings()
	if err != nil {
		log.Fatalf("error reading settings.json, %v", err)
	}
	clientTukuiAddons := getExtraUIAddons("elvui", "tukui")

	if installAddon != "" {
		install(settings, clientTukuiAddons)
		return
	}

	tukuiAddons := getTukuiAddonList()
	addonsFolder := filepath.Join(settings.WowDirectory, "interface", "addons")
	addons := getLocalTukuiAddons(addonsFolder)

	wg := sync.WaitGroup{}
	wg.Add(len(addons))

	for _, addon := range addons {
		go checkForUpdate(addon, tukuiAddons, clientTukuiAddons, addonsFolder, &wg)
	}
	wg.Wait()

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
	log.Infof("%s downloading", name)
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
	log.Infof("%s unzipping", zipname)
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
		log.Warn(installAddon, "no supported yet...")
		return
	}
	a := addons[addon]
	zipname := fmt.Sprintf("%s.zip", addonZipFileName.ReplaceAllString(a.Name, ""))
	defer os.Remove(zipname)
	err := downloadAddon(a.URL, zipname)
	if err != nil {
		log.Error(err)
		return
	}

	_, err = Unzip(zipname, addonsFolder)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("done")
}

func checkForUpdate(addon LocalTukuiAddon, tukuiAddons []TukuiAddon, clientTukuiAddons []ClientApiAddon, addonsFolder string, wg *sync.WaitGroup) {
	defer wg.Done()
	ver := ""
	downloadUrl := ""

	if addon.Toc.XTukuiProjectID < 0 {
		clienttukuiaddon := getClientAddonbyID(clientTukuiAddons, addon.Toc.XTukuiProjectID)
		if clienttukuiaddon != nil {
			ver = clienttukuiaddon.Version
			downloadUrl = clienttukuiaddon.URL
		}
	} else {
		tukuiaddon := getAddonbyID(tukuiAddons, fmt.Sprintf("%d", addon.Toc.XTukuiProjectID))
		if tukuiaddon != nil {
			ver = tukuiaddon.Version
			downloadUrl = tukuiaddon.URL
		}
	}

	if ver == "" || downloadUrl == "" {
		log.Infof("%s couldn't find the addon in the api", addon.Name)
		return
	}

	addonVersion, err := version.NewVersion(ver)
	if err != nil {
		log.Error(err)
		return
	}
	if addonVersion.GreaterThan(addon.Toc.Version) {
		log.Infof("updating %s", addon.Name)
		err = addon.Update(downloadUrl, addonsFolder)
		if err != nil {
			log.Errorf("error updating %s, error: %v", addon.Name, err)
			return
		}
		updateToc(addon.Toc.path, addon.Toc.XTukuiProjectID, addonVersion.String())
		log.Infof("finished updating %s from %s to %s", addon.Name, addon.Toc.Version, addonVersion)
		return
	}
	log.Infof("%s is up-to-date!", addon.Name)
}
