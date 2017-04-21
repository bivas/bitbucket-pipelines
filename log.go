package main

import (
	"io/ioutil"
	"log"
	"os"
)

func logSetup() {
	temp, e := ioutil.TempFile("", "bitbucket-pipeline-log.")
	if e != nil {
		panic(e)
	}
	log.SetOutput(os.Stderr)
	log.Println("Log file at", temp.Name())
	log.SetOutput(temp)
}
