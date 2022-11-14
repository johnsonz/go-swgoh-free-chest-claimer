package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	dateLayout = "2006-01-02"
)

func main() {
	for {
		time.Sleep(time.Minute * 5)
		if time.Now().Hour() == 19 {
			for _, player := range config.Players {
				if player.Skip || player.LastClaimedDate == time.Now().Format(dateLayout) {
					continue
				}

				log.Println(strings.Repeat("*", 100))
				log.Println(player.Nickname, player.Email)
				var content string
				if ok, item, msg := claim(player); ok {
					content = fmt.Sprintf(contentSucceeded, player.Nickname, item, time.Now().Format(datetimeLayout))
				} else {
					content = fmt.Sprintf(contentFalied, player.Nickname, item, msg, time.Now().Format(datetimeLayout))
				}
				sendDiscordMessage(content)
			}
		}
	}
}
