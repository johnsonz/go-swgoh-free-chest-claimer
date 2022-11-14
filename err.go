package main

import (
	"github.com/fatih/color"
	"github.com/golang/glog"
)

// custom log level
const (
	Info = iota
	Warning
	Debug
	Error
)

func checkErr(messge string, err error, level int) {
	if err != nil {
		switch level {
		case Info:
			color.Set(color.FgGreen)
			defer color.Unset()
			glog.Infoln(messge, err)
		case Warning, Debug:
			glog.Infoln(messge, err)
		case Error:
			glog.Fatalln(messge, err)
		}
	}
}
