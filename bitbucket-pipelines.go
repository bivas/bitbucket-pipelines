package main

import (
	"github.com/bivas/bitbucket-pipelines/runner"
	"log"
	"os"
)

func init() {
	logSetup()
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Unable to access current directory", err)
	}
	runner.Run(wd + "/bitbucket-pipelines.yml")
}
