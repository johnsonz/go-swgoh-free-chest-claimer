package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net/mail"
	"os"
)

const _MAX_LINE_LEN = 1024

var crlf = []byte{'\r', '\n'}

// If debug is true, errors parsing messages will be printed to stderr. If
// false, they will be ignored. Either way those messages will not appear in
// the msgs slice.
func readMail(r io.Reader, debug bool) (msgs []*mail.Message, err error) {
	var mbuf *bytes.Buffer
	lastblank := true
	br := bufio.NewReaderSize(r, _MAX_LINE_LEN)
	l, _, err := br.ReadLine()
	for err == nil {
		fs := bytes.SplitN(l, []byte{' '}, 3)
		if len(fs) == 3 && string(fs[0]) == "From" && lastblank {
			// flush the previous message, if necessary
			if mbuf != nil {
				msgs = parseAndAppendMail(mbuf, msgs, debug)
			}
			mbuf = new(bytes.Buffer)
		} else {
			_, err = mbuf.Write(l)
			if err != nil {
				return
			}
			_, err = mbuf.Write(crlf)
			if err != nil {
				return
			}
		}
		if len(l) > 0 {
			lastblank = false
		} else {
			lastblank = true
		}
		l, _, err = br.ReadLine()
	}
	if err == io.EOF {
		msgs = parseAndAppendMail(mbuf, msgs, debug)
		err = nil
	}
	return
}

// If debug is true, errors parsing messages will be printed to stderr. If
// false, they will be ignored. Either way those messages will not appear in
// the msgs slice.
func readMailFile(filename string, debug bool) ([]*mail.Message, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	msgs, err := readMail(f, debug)
	f.Close()
	return msgs, err
}

func parseAndAppendMail(mbuf *bytes.Buffer, msgs []*mail.Message, debug bool) []*mail.Message {
	msg, err := mail.ReadMessage(mbuf)
	if err != nil {
		if debug {
			log.Print(err)
		}
		return msgs // don't append
	}
	return append(msgs, msg)
}
