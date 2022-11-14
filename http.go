package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
)

func getHttpResp(url string) string {
	req, err := http.NewRequest("GET", url, nil)
	checkErr("create Http request error: ", err, Error)

	client := &http.Client{Transport: &http.Transport{
		IdleConnTimeout: 180 * time.Second,
	}}
	resp, err := client.Do(req)
	checkErr("fetch Http request error: ", err, Error)

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	checkErr("read Http request body error: ", err, Error)

	return string(b)
}
func postReq(url, method string, content []byte, header http.Header) (http.Header, []byte) {
	var req *http.Request
	var err error
	if strings.ToUpper(method) == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(content))
	}
	checkErr("create Http request error: ", err, Error)

	header.Set("content-type", "application/json")
	header.Set("accept-language", "en")
	header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36")
	req.Header = header
	client := &http.Client{}
	resp, err := client.Do(req)
	checkErr("create Http request error: ", err, Error)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	checkErr("read Http request body error: ", err, Error)

	return resp.Header, body
}
