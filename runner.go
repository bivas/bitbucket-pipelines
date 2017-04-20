package bitbucketpipelines

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
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
       - docker version
options:
 docker: true
`

const (
	pipelineRunnerName       = "pipeline__runner__"
	pipelineRunnerDockerName = pipelineRunnerName + "docker"
	dockerImage              = "docker:1.8-dind"
	bootTimeout              = 30
)

//PipelineRunner : create the pipelines runner
type PipelineRunner interface {
	Setup()
	Run(commands []string) (string, error)
	Close()
}

type runner struct {
	image       string
	hostPath    string
	dockerMount bool
	command     *exec.Cmd
	output      bytes.Buffer
	signal      chan error
}

func (runner) commandRun(name string, args []string) *exec.Cmd {
	log.Println("Running (commandRun)", name, args)
	return exec.Command(name, args...)
}

func (runner) commandOutput(name string, args []string) (string, error) {
	log.Println("Running (commandOutput)", name, args)
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
		log.Fatalf("Error pulling image %s %s\n", env.image, e)
		log.Fatalf("Error message %s\n", out)
	}
}

func (env runner) runDockerDocker() {
	if env.dockerMount {
		log.Println("pulling", dockerImage)
		out, e := env.docker("pull", dockerImage)
		if e != nil {
			log.Fatalf("Error pulling image %s %s\n", dockerImage, e)
			log.Fatalf("Error message %s\n", out)
		}
		args := []string{"run",
			"-d",
			"--name=" + pipelineRunnerDockerName,
			"--privileged",
			dockerImage}
		dockerCommand := env.commandRun("docker", args)
		log.Println("running", dockerImage)
		go func() {
			out, err := dockerCommand.CombinedOutput()
			if err != nil {
				log.Fatalf("error combining output %s, %s", out, err)
			}
		}()
		env.waitForImage(pipelineRunnerDockerName)
	}
}

func (env runner) sendInitCommands(cmd *exec.Cmd) {
	stdin, e := cmd.StdinPipe()
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
		"--entrypoint=/bin/sh"}
	if env.dockerMount {
		env.runDockerDocker()
		args = append(args,
			[]string{
				"--env=DOCKER_HOST=tcp://docker:2375",
				"--link=" + pipelineRunnerDockerName + ":docker",
			}...)
	}
	args = append(args, env.image)
	env.command = env.commandRun("docker", args)
	env.sendInitCommands(env.command)
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
	if env.dockerMount {
		env.docker("stop", pipelineRunnerDockerName)
	}
}

func (env runner) cleanup() {
	env.docker("rm", "-f", pipelineRunnerName)
	if env.dockerMount {
		env.docker("rm", "-f", pipelineRunnerDockerName)
		os.Remove("")
	}
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

func (env runner) waitForImage(image string) error {
	filterPs := []string{"ps", "-aq", "--filter", "name=" + image}
	id, _ := env.docker(filterPs...)
	for i := 0; ; i++ {
		if i > bootTimeout {
			return fmt.Errorf("Unable to start container after %d seconds", bootTimeout)
		}
		log.Println("Waiting for container to be available", image, id)
		if id != "" {
			return nil
		}
		time.Sleep(1 * time.Second)
		id, _ = env.docker(filterPs...)
	}
	return nil
}

func (env *runner) Run(commands []string) (string, error) {
	go func() {
		time.Sleep(5 * time.Minute)
		env.signal <- errors.New("Timeout trying to run commands")
	}()
	go func() {
		e := env.waitForImage(pipelineRunnerName)
		if e != nil {
			env.signal <- e
			return
		}
		if env.dockerMount {
			if out1, e1 := env.docker([]string{
				"cp",
				pipelineRunnerDockerName + ":/usr/local/bin/docker",
				"/tmp/",
			}...); e1 != nil {
				log.Fatal("copy from", out1, e1)
			}
			if out1, e1 := env.docker([]string{
				"exec",
				pipelineRunnerName,
				"mkdir",
				"-p",
				"/usr/local/bin/",
			}...); e1 != nil {
				log.Fatal("copy to", out1, e1)
			}
			if out1, e1 := env.docker([]string{
				"cp",
				"/tmp/docker",
				pipelineRunnerName + ":/usr/local/bin/",
			}...); e1 != nil {
				log.Fatal("copy to", out1, e1)
			}
		}
		for _, command := range commands {
			output, err := env.docker("exec", "-i", pipelineRunnerName, "/bin/sh", "-c", command)
			if err != nil {
				env.signal <- errors.New(fmt.Sprintln("error running", command, output, err))
			}
			env.output.WriteString(fmt.Sprintf("\n == Running '%s' ==>\n", command))
			env.output.WriteString(output)
			env.output.WriteByte('\n')
		}
		env.signal <- nil
	}()
	err := <-env.signal
	env.stopImage()
	return string(env.output.Bytes()), err
}

//Run : run it!
func Run() {
	reader := strings.NewReader(data)
	pipline := ReadPipelineDef(reader)
	path, _ := os.Getwd()
	env := &runner{
		image:       pipline.Image,
		hostPath:    path,
		dockerMount: pipline.Options.Docker,
	}
	defer env.Close()
	env.Setup()
	output, e := env.Run(pipline.Pipelines.Default[0].Step.Scripts)
	if e != nil {
		log.Fatal(output, e)
	}
	fmt.Println(output)
}
