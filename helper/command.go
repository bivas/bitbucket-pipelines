package helper

import (
	"io"
	"log"
	"os/exec"
	"strings"
)

type Command interface {
	Run() (string, error)
	Stdin() io.WriteCloser
	Kill()
}

type command struct {
	name  string
	args  []string
	under *exec.Cmd
}

func (c *command) Stdin() io.WriteCloser {
	pipe, e := c.under.StdinPipe()
	if e != nil {
		log.Fatal("error setting up stdin", e)
	}
	return pipe
}

func (c *command) Kill() {
	c.under.Process.Kill()
}

func (c *command) Run() (string, error) {
	log.Println("Running", c.name, c.args)
	out, err := c.under.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func NewCommand(name string, args []string) Command {
	return &command{
		name:  name,
		args:  args,
		under: exec.Command(name, args...),
	}
}
