package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type Toc struct {
	Version              float64
	XTukuiProjectID      int
	XTukuiProjectFolders string
	lines                []string
	path                 string
}

func parseToc(path string) *Toc {
	toc := Toc{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
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
			version, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				log.Println(err)
				continue
			}
			toc.Version = version
		case "X-Tukui-ProjectID":
			id, err := strconv.Atoi(parts[2])
			if err != nil {
				log.Println(err)
				continue
			}
			toc.XTukuiProjectID = id
		case "X-Tukui-ProjectFolders":
			toc.XTukuiProjectFolders = strings.TrimSpace(parts[2])
		default:
		}
	}
	return &toc
}

func updateToc(path string, id int) error {
	newtoc := parseToc(path)
	endstring := bytes.NewBuffer([]byte{})
	hasId := false
	lastHash := 0
	for i, line := range newtoc.lines {
		if strings.Contains(line, "## X-Tukui-ProjectID: ") {
			hasId = true
			break
		}
		if !strings.Contains(line, "##") {
			lastHash = i
			break
		}
	}
	if hasId {
		return nil
	}

	newtoc.lines = append(newtoc.lines, "")
	copy(newtoc.lines[lastHash+1:], newtoc.lines[lastHash:])
	newtoc.lines[lastHash] = fmt.Sprintf("## X-Tukui-ProjectID: %d", id)

	for _, line := range newtoc.lines {
		endstring.WriteString(fmt.Sprintf("%s\r\n", strings.TrimSpace(line)))
	}

	return ioutil.WriteFile(path, endstring.Bytes(), 0755)
}
