package bitbucketpipelines

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
	"errors"
)

var data = `
image: python:alpine
pipelines:
 default:
  - step:
     script:
       - ls
       - ps
       - python --version
`

const (
	pipelineRunnerName = "pipeline__runner__"
	bootTimeout = 30
)

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
	output   bytes.Buffer
	signal   chan error
}

func (runner) commandRun(name string, args []string) *exec.Cmd {
	return exec.Command(name, args...)
}

func (runner) commandOutput(name string, args []string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (env runner) docker(args ...string) (string, error) {
	return env.commandOutput("docker", args)
}

func (env runner) pullImage() {
	log.Println("pulling image", env.image)
	out, e := env.docker("pull", env.image)
	if e != nil {
		log.Fatal("Error pulling image", env.image, e)
		log.Fatal("Error message", out)
	}
}

func (env runner) sendInitCommands() {
	stdin, e := env.command.StdinPipe()
	if e != nil {
		log.Fatal("error setting up stdin", e)
	}
	defer stdin.Close()
	io.WriteString(stdin, "touch /.running\n")
	io.WriteString(stdin, "while [ -e /.running ]; do sleep 1; done; exit;\n")
}

func (env *runner) runImage() {
	env.cleanup()
	log.Println("running image", env.image)
	args := []string{"run",
		"-i",
		"--rm",
		"--name=" + pipelineRunnerName,
		"--volume=" + env.hostPath + ":/wd",
		"--workdir=/wd",
		"--entrypoint=/bin/sh",
		env.image}
	env.command = env.commandRun("docker", args)
	env.sendInitCommands()
	env.signal = make(chan error)
	go func() {
		out, err := env.command.CombinedOutput()
		if err != nil {
			env.signal <- fmt.Errorf("error combining output %s, %s", out, err)
		} else {
			env.signal <- err
		}
	}()
}

func (env runner) stopImage() {
	env.docker("exec", "-i", pipelineRunnerName, "rm", "/.running")
}

func (env runner) cleanup() {
	env.docker("rm", "-f", pipelineRunnerName)
}

func (env *runner) Setup() {
	log.Println("Setup runner")
	env.pullImage()
	env.runImage()
}

func (env *runner) Close() {
	log.Println("Closing runner")
	env.command.Process.Kill()
	env.command.Wait()
	env.cleanup()
}

func (env runner) waitForImage() {
	filterPs := []string{"ps", "-aq", "--filter", "name=" + pipelineRunnerName}
	id, _ := env.docker(filterPs...)
	for i := 0; ; i++ {
		if i > bootTimeout {
			env.signal <- fmt.Errorf("Unable to start container after %d seconds", bootTimeout)
		}
		log.Println("Waiting for container to be available", id)
		if id != "" {
			break
		}
		time.Sleep(1 * time.Second)
		id, _ = env.docker(filterPs...)
	}
}

func (env *runner) Run(commands []string) (string, error) {
	go func() {
		time.Sleep(5 * time.Minute)
		env.signal <- errors.New("Timeout trying to run commands")
	}()
	go func() {
		env.waitForImage()
		for _, command := range commands {
			output, err := env.docker("exec", "-i", pipelineRunnerName, "/bin/sh", "-c", command)
			if err != nil {
				env.signal <- errors.New(fmt.Sprintln("error running", command, output, err))
			}
			env.output.WriteString(fmt.Sprintf("\n == Running '%s' ==>\n", command))
			env.output.WriteString(output)
			env.output.WriteByte('\n')
		}
		env.stopImage()
	}()
	err := <-env.signal
	return string(env.output.Bytes()), err
}

//Run : run it!
func Run() {
	reader := strings.NewReader(data)
	pipline := ReadPipelineDef(reader)
	path, _ := os.Getwd()
	env := &runner{
		image:    pipline.Image,
		hostPath: path,
	}
	defer env.Close()
	env.Setup()
	output, e := env.Run(pipline.Pipelines.Default[0].Step.Scripts)
	if e != nil {
		log.Fatal(output, e)
	}
	fmt.Println(output)
}
