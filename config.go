package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
)

type Config struct {
	Players []Player `json:"Players"`
	Discord Discord  `json:"Discord"`
	Mbox    Mbox     `json:"Mbox"`
}
type Player struct {
	Nickname        string `json:"Nickname"`
	Email           string `json:"Email"`
	DiscordId       string `json:"DiscordId"`
	Skip            bool   `json:"Skip"`
	LastClaimedDate string
}
type Discord struct {
	Token            string `json:"Token"`
	Guild            string `json:"Guild"`
	RemoveCommands   bool   `json:"RemoveCommands"`
	AutoMsgToChannel string `json:"AutoMsgToChannel"`
}
type Mbox struct {
	Path string `json:"Path"`
}

const (
	configFileName string = "config.json"
)

var config Config

func init() {
	config = parseConfig()
}
func parseConfig() Config {
	conf, err := os.ReadFile(configFileName)
	checkErr("read config file error: ", err, Error)

	var lines []string
	for _, line := range strings.Split(strings.Replace(string(conf), "\r\n", "\n", -1), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "//") && line != "" {
			lines = append(lines, line)
		}
	}

	var b bytes.Buffer
	for i, line := range lines {
		if len(lines)-1 > i {
			nextLine := lines[i+1]
			if nextLine == "]" || nextLine == "]," || nextLine == "}" || nextLine == "}," {
				line = strings.TrimSuffix(line, ",")
			}
		}
		b.WriteString(line)
	}
	var config Config
	err = json.Unmarshal(b.Bytes(), &config)
	checkErr("parse config file error: ", err, Error)
	return config
}
