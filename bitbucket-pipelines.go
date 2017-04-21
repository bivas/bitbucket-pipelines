package main

import (
	"flag"
	"github.com/bivas/bitbucket-pipelines/runner"
	"log"
	"os"
)

func main() {
	yamlFile := flag.String("yaml", "bitbucket-pipelines.yml", "Specify pipelines yaml file")
	envFile := flag.String("env", "", "Add environment variables to pipeline")
	flag.Parse()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Unable to access current directory", err)
	}
	logSetup()
	runner.Run(wd+"/"+*yamlFile, *envFile)
}
