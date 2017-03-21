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
)

var data = `
image: busybox
pipelines:
 default:
  - step:
     script:
       - ls
       - ps
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
	output   bytes.Buffer
}

func (runner) commandRun(name string, args []string) *exec.Cmd {
	log.Println("running command", name, args)
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
		"-i",
		"--name=pipeline__runner__",
		"--volume=" + env.hostPath + ":/repo",
		"--workdir=/repo",
		"--entrypoint=/bin/sh",
		env.image}
	env.command = env.commandRun("docker", args)
	log.Printf("%+v\n", env.command)
	var e error
	env.stdin, e = env.command.StdinPipe()
	if e != nil {
		log.Fatal(e)
	}
	log.Println("setting up pipes")
	log.Printf("%+v\n", env.command)
}

func (env *runner) Setup() {
	log.Println("setup runner")
	//env.pullImage()
	env.runImage()
}

func (env *runner) Close() {
	log.Println("closing running")
	env.stdin.Close()
	env.command.Process.Kill()
	env.command.Wait()
}

func (env *runner) Run(commands []string) (string, error) {
	go func() {
		id, _ := env.commandOutput("docker", []string{"ps", "-aq", "--filter", "name=pipeline__runner__"})
		for {
			log.Println("Waiting for container to be available", id)
			if id != "" {
				break
			}
			time.Sleep(750 * time.Millisecond)
			id, _ = env.commandOutput("docker", []string{"ps", "-aq", "--filter", "name=pipeline__runner__"})
		}
		for _, command := range commands {
			output, err := env.commandOutput("docker", []string{"exec", "-i", "pipeline__runner__", command})
			if err != nil {
				log.Fatalln(output, err)
			}
			log.Println("command", command, "output", output)
			env.output.WriteString(fmt.Sprintf("\n##### Running '%s' ==>\n", command))
			env.output.WriteString(output)
			env.commandOutput("docker", []string{"exec", "-i", "pipeline__runner__", "rm", "/.running"})
		}

	}()
	defer env.stdin.Close()
	io.WriteString(env.stdin, "touch /.running\n")
	io.WriteString(env.stdin, "while [ -e /.running ]; do sleep 1; done; exit;\n")
	out, err := env.command.CombinedOutput()
	if err != nil {
		log.Fatal(out, err)
	}
	return string(env.output.Bytes()), err
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
	defer env.Close()
	env.Setup()
	output, e := env.Run(pipline.Pipelines.Default[0].Step.Scripts)
	if e != nil {
		log.Fatal(output, e)
	}
	log.Println("output", output)
}
