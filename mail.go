package main

import (
	"strings"
	"time"
)

const (
	longDatetimeLayout = "Mon, 02 Jan 2006 15:04:05 -0700"
)

func getCodeFromEmail(to string) string {
	mails, err := readMailFile(config.Mbox.Path, false)
	checkErr("read email file error: ", err, Error)
	for i := len(mails) - 1; i > 0; i-- {
		header := mails[i].Header
		subject := header.Get("Subject")
		if strings.Contains(subject, "Verification Code For EA -") {
			toUser := header.Get("To")
			if strings.Contains(strings.ToLower(toUser), to) {
				from := header.Get("From")
				if strings.Contains(strings.ToLower(from), "ea@e.ea.com") {
					date := header.Get("Date")

					loc, _ := time.LoadLocation("Asia/Shanghai")
					t, err := time.ParseInLocation(longDatetimeLayout, date, loc)
					checkErr("parse time from email header error", err, Info)
					if err != nil {
						break
					}
					if t.Add(time.Minute * 5).After(time.Now()) {
						code := strings.TrimSpace(subject[strings.LastIndex(subject, "-")+1:])
						return code
					}
				}
			}
		}
	}
	return ""
}
