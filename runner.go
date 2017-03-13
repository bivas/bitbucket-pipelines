package bitbucketpipelines

import (
	"os/exec"
	"strings"
)

var data = `
image: python:2.7
pipelines:
default:
 - step:
    script:
      - python --version
      - python myScript.py
`

func commandRun(name string, args []string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func pullImage(image string) {
	args := []string{"pull", image}
	commandRun("docker", args)
}

func runImage(image string) {
	args := []string{"run",
		"-it",
		"--rm",
		"--volume=/Users/myUserName/code/localDebugRepo:/repo",
		"--workdir='/repo'",
		"--memory=4g",
		"--entrypoint=/bin/bash",
		image}
	commandRun("docker", args)
}

//Run : run it!
func Run() {
	reader := strings.NewReader(data)
	pipline := ReadPipelineDef(reader)
	pullImage(pipline.Image)
}
