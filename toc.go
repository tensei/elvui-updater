package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

// Toc ...
type Toc struct {
	Version              *version.Version
	XTukuiProjectID      int
	HasProjectID         bool
	XTukuiProjectFolders string

	lines []string
	path  string
}

func parseToc(path string) *Toc {
	toc := Toc{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		return &toc
	}
	toc.path = path
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		toc.lines = append(toc.lines, line)
		parts := strings.SplitN(line, " ", 3)
		if parts[0] != "##" {
			continue
		}
		tag := parts[1][:len(parts[1])-1]
		switch tag {
		case "Version":
			toc.Version, err = version.NewVersion(parts[2])
			if err != nil {
				log.Warnf("error parsing version, %w", err)
				continue
			}
		case "X-Tukui-ProjectID":
			id, err := strconv.Atoi(parts[2])
			if err != nil {
				log.Warnf("error parsing ProjectID to int, %w", err)
				continue
			}
			toc.XTukuiProjectID = id
			toc.HasProjectID = true
		case "X-Tukui-ProjectFolders":
			toc.XTukuiProjectFolders = strings.TrimSpace(parts[2])
		default:
		}
	}
	return &toc
}

func updateToc(path string, id int, version string) error {
	newtoc := parseToc(path)
	endstring := bytes.NewBuffer([]byte{})
	hasID := false
	lastHash := 0
	for i, line := range newtoc.lines {
		if strings.Contains(line, "## X-Tukui-ProjectID: ") {
			hasID = true
			break
		}
		if !strings.Contains(line, "##") {
			lastHash = i
			break
		}
	}
	if hasID {
		return nil
	}

	newtoc.lines = append(newtoc.lines, "")
	copy(newtoc.lines[lastHash+1:], newtoc.lines[lastHash:])
	newtoc.lines[lastHash] = fmt.Sprintf("## X-Tukui-ProjectID: %d", id)

	for _, line := range newtoc.lines {
		if strings.HasPrefix(line, "## Version: ") && version != "" {
			line = "## Version: " + strings.TrimSpace(version)
		}
		endstring.WriteString(fmt.Sprintf("%s\r\n", strings.TrimSpace(line)))
	}

	return ioutil.WriteFile(path, endstring.Bytes(), 0755)
}
