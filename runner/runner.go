package runner

import (
	"errors"
	"fmt"
	"github.com/bivas/bitbucket-pipelines/helper"
	"github.com/bivas/bitbucket-pipelines/parser"
	"github.com/bivas/bitbucket-pipelines/service"
	"github.com/bivas/bitbucket-pipelines/service/docker"
	"io"
	"log"
	"os"
	"time"
)

const (
	pipelineRunnerName = "pipeline__runner__"
)

//PipelineRunner : create the pipelines inner
type PipelineRunner interface {
	Setup()
	Run(helpers []string) (string, error)
	Close()
}

type inner struct {
	image            string
	hostPath         string
	environment      string
	useDockerService bool
	helper           helper.Command
	signal           chan error
	services         []service.Service
}

func (env *inner) pullImage() {
	ui.Info("Pulling image %s", env.image)
	out, e := helper.NewDockerCommand("pull", env.image).Run()
	if e != nil {
		ui.Error("Error pulling image %s %s", env.image, e)
		log.Fatalf("Error message %s\n", out)
	}
}

func (*inner) sendInitCommands(cmd helper.Command) {
	stdin := cmd.Stdin()
	defer stdin.Close()
	io.WriteString(stdin, "touch /.running\n")
	io.WriteString(stdin, "while [ -e /.running ]; do sleep 1; done; exit;\n")
}

func (env *inner) runImage() {
	env.cleanup()
	ui.Info("Running image %s", env.image)
	args := []string{"run",
		"-i",
		"--rm",
		"--name=" + pipelineRunnerName,
		"--volume=" + env.hostPath + ":/wd",
		"--workdir=/wd",
		"--entrypoint=/bin/sh",
		"--env=CI=true",
	}
	if hash, e := helper.LatestCommitHash(); e == nil {
		args = append(args, "--env=BITBUCKET_COMMIT="+hash)
	}
	for _, svc := range env.services {
		svcArgs, e := svc.Start()
		if e != nil {
			ui.Error("Unable to start service %v\n%s", svc, e)
		}
		args = append(args, svcArgs...)
	}
	if env.environment != "" {
		if helper.IsFile(env.environment) {
			args = append(args,
				[]string{
					"--env-file=" + env.environment,
				}...)
		} else {
			ui.Warning(
				"Environment file %s doesn't exist - running without environment overrides",
				env.environment)
		}
	}
	args = append(args, env.image)
	env.helper = helper.NewDockerCommand(args...)
	env.sendInitCommands(env.helper)
	env.signal = make(chan error)
	go func() {
		out, err := env.helper.Run()
		if err != nil {
			env.signal <- fmt.Errorf("error combining output %s, %s", out, err)
		} else {
			env.signal <- err
		}
	}()
}

func (env *inner) stopImage() {
	helper.NewDockerCommand("exec", "-i", pipelineRunnerName, "rm", "/.running").Run()
}

func (env *inner) cleanup() {
	helper.NewDockerCommand("rm", "-f", pipelineRunnerName).Run()
	for _, svc := range env.services {
		svc.Stop()
	}
}

func (env *inner) Setup() {
	log.Println("Setup inner runner")
	env.initServices()
	env.pullImage()
	env.runImage()
}
func (env *inner) initServices() {
	if env.useDockerService {
		env.services = append(env.services, docker.NewService())
	}
	for _, svc := range env.services {
		svc.Init()
	}
}

func (env *inner) Close() {
	log.Println("Closing inner runner")
	env.helper.Kill()
	env.cleanup()
}

func (env *inner) Run(helpers []string) error {
	go func() {
		time.Sleep(5 * time.Minute)
		env.signal <- errors.New("Timeout trying to run helpers")
	}()
	go func() {
		e := helper.WaitForContainer(pipelineRunnerName)
		if e != nil {
			env.signal <- e
			return
		}
		for _, svc := range env.services {
			svc.Attach(pipelineRunnerName)
		}
		for _, current := range helpers {
			output, err := helper.NewDockerCommand("exec", "-i", pipelineRunnerName, "/bin/sh", "-c", current).Run()
			if err != nil {
				ui.Error(" == Running '%s' ==>", current)
				ui.Output(output)
				env.signal <- errors.New(fmt.Sprintln("helper", current, output, err))
				return
			}
			ui.Info(" == Running '%s' ==>", current)
			ui.Output(output)
		}
		env.signal <- nil
	}()
	err := <-env.signal
	env.stopImage()
	return err
}

//Run : run it!
func Run(yml string, env string) int {
	reader, e := os.Open(yml)
	if e != nil {
		log.Printf("Unable to read %s\n", yml)
		return 1
	}
	pipeline := parser.ReadPipelineDef(reader)
	path, _ := os.Getwd()
	r := &inner{
		image:            pipeline.Image,
		hostPath:         path,
		environment:      env,
		useDockerService: pipeline.Options.Docker,
	}
	defer r.Close()
	r.Setup()
	eR := r.Run(pipeline.Pipelines.Default[0].Step.Scripts)
	if eR != nil {
		log.Println("error when running", eR)
		return 1
	}
	return 0
}
