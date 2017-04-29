package helper

import (
	"fmt"
	"log"
	"time"
)

const (
	bootTimeout = 30
)

func NewDockerCommand(args ...string) Command {
	return NewCommand("docker", args)
}

func WaitForContainer(container string) error {
	filterPs := []string{"ps", "-aq", "--filter", "name=" + container}
	id, _ := NewDockerCommand(filterPs...).Run()
	for i := 0; ; i++ {
		if i > bootTimeout {
			return fmt.Errorf("Unable to start container after %d seconds", bootTimeout)
		}
		log.Println("Waiting for container to be available", container, id)
		time.Sleep(1 * time.Second)
		if id != "" {
			return nil
		}
		id, _ = NewDockerCommand(filterPs...).Run()
	}
	return nil
}
