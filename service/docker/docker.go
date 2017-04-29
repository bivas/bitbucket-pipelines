package docker

import (
	"fmt"
	"github.com/bivas/bitbucket-pipelines/helper"
	"github.com/bivas/bitbucket-pipelines/service"
	"log"
)

const (
	pipelineServiceDockerName = service.PipelineServiceName + "docker"
	dockerImage               = "docker:17-dind"
)

type DockerService struct {
	command helper.Command
	err     error
}

func (*DockerService) pullImage() {
	log.Println("[DockerService] pulling", dockerImage)
	out, e := helper.NewDockerCommand("pull", dockerImage).Run()
	if e != nil {
		log.Fatalf("Error pulling image %s %s\n", dockerImage, e)
		log.Fatalf("Error message %s\n", out)
	}
}

func (s *DockerService) Init() {
	s.pullImage()
	args := []string{"run",
		"-d",
		"--name=" + pipelineServiceDockerName,
		"--privileged",
		dockerImage}
	s.command = helper.NewDockerCommand(args...)
}

func (s *DockerService) Start() ([]string, error) {
	log.Println("[DockerService] running", dockerImage)
	go func() {
		out, err := s.command.Run()
		if err != nil {
			s.err = fmt.Errorf("error combining output %s, %s", out, err)
		}
	}()
	if e := helper.WaitForContainer(pipelineServiceDockerName); e != nil {
		s.err = e
	}
	if s.err != nil {
		return []string{}, s.err
	} else {
		return []string{
			"--env=DOCKER_HOST=tcp://docker:2375",
			"--link=" + pipelineServiceDockerName + ":docker",
		}, nil
	}
}

func (*DockerService) Attach(container string) error {
	if out1, e1 := helper.NewDockerCommand(
		"cp",
		pipelineServiceDockerName+":/usr/local/bin/docker",
		"/tmp/",
	).Run(); e1 != nil {
		return fmt.Errorf("[DockerService] copy from %s %s", out1, e1)
	}
	if out1, e1 := helper.NewDockerCommand(
		"exec",
		container,
		"mkdir",
		"-p",
		"/usr/local/bin/",
	).Run(); e1 != nil {
		return fmt.Errorf("[DockerService] copy from %s %s", out1, e1)
	}
	if out1, e1 := helper.NewDockerCommand(
		"cp",
		"/tmp/docker",
		container+":/usr/local/bin/",
	).Run(); e1 != nil {
		return fmt.Errorf("[DockerService] copy from %s %s", out1, e1)
	}
	return nil
}

func (*DockerService) Stop() {
	helper.NewDockerCommand("stop", pipelineServiceDockerName).Run()
	helper.NewDockerCommand("rm", "-f", pipelineServiceDockerName).Run()
}

func NewService() service.Service {
	return &DockerService{}
}
