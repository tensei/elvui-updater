package main

import (
	"log"
	"strconv"
	"strings"
)

type Toc struct {
	Version              float64
	XTukuiProjectID      int
	XTukuiProjectFolders []string
}

func parseToc(data string) *Toc {
	toc := Toc{}
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), " ")
		if parts[0] != "##" || len(parts) < 3 {
			continue
		}
		switch parts[1][:len(parts[1])-1] {
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
			toc.XTukuiProjectFolders = strings.Split(parts[2], ",")
		default:
		}
	}
	return &toc
}
