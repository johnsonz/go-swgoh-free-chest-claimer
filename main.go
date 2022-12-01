package main

import (
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
			for i, player := range config.Players {
				if player.Skip || player.LastClaimedDate == time.Now().Format(dateLayout) {
					continue
				}
				config.Players[i].LastClaimedDate = time.Now().Format(dateLayout)
				log.Println(strings.Repeat("*", 100))
				log.Println(player.Nickname, player.Email)
				var content string
				items := claim(player)
				content = parseItemClaimMessage(items, player.Nickname)
				sendDiscordMessage(content)
			}
		}
	}
}
