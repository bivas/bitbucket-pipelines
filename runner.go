package bitbucketpipelines

import (
	"bytes"
	"io"
	"log"
	"os"
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

//PipelineRunner : create the pipelines runner
type PipelineRunner interface {
	Setup()
	Run(commands []string) (string, error)
	Close()
}

type runner struct {
	image    string
	hostPath string
	command  *exec.Cmd
	stdin    io.WriteCloser
	stdout   io.ReadCloser
	stderr   io.ReadCloser
}

func (runner) commandRun(name string, args []string) *exec.Cmd {
	return exec.Command(name, args...)
}

func (runner) commandOutput(name string, args []string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (env runner) pullImage() {
	log.Println("pulling image", env.image)
	args := []string{"pull", env.image}
	env.commandOutput("docker", args)
	log.Println("pulled image", env.image)
}

func (env *runner) runImage() {
	log.Println("running image", env.image)
	args := []string{"run",
		"-it",
		//"--rm",
		"--volume=" + env.hostPath + ":/repo",
		"--workdir='/repo'",
		"--memory=4g",
		"--entrypoint=/bin/bash",
		env.image}
	env.command = env.commandRun("docker", args)
	var e error
	env.stdin, e = env.command.StdinPipe()
	if e != nil {
		log.Fatal(e)
	}
	env.stdout, e = env.command.StdoutPipe()
	if e != nil {
		log.Fatal(e)
	}
	env.stderr, e = env.command.StderrPipe()
	if e != nil {
		log.Fatal(e)
	}
	log.Println("setting up pipes")
	if e := env.command.Start(); e != nil {
		log.Fatal("error running docker", e)
	}
}

func (env runner) Setup() {
	log.Println("setup runner")
	env.pullImage()
	env.runImage()
}

func (env *runner) Close() {
	log.Println("closing running")
	env.stdin.Close()
	env.stdout.Close()
	env.stderr.Close()
	env.command.Process.Kill()
	env.command.Wait()
}

func (env *runner) Run(commands []string) (string, error) {
	for _, command := range commands {
		log.Println("running command", command)
		io.WriteString(env.stdin, command+"/n")
	}
	var b bytes.Buffer
	env.command.Stdout = &b
	env.command.Stderr = &b
	return string(b.Bytes()), nil
}

//Run : run it!
func Run() {
	reader := strings.NewReader(data)
	pipline := ReadPipelineDef(reader)
	log.Printf("%+v", pipline)
	path, _ := os.Getwd()
	env := &runner{
		image:    pipline.Image,
		hostPath: path,
	}
	log.Printf("%+v", env)
	defer env.Close()
	env.Setup()
	output, e := env.Run(pipline.Pipelines.Default[0].Step.Scripts)
	if e != nil {
		log.Fatal(e)
	}
	log.Println("output", output)
}
