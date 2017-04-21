package runner

import (
	"fmt"
	"github.com/fatih/color"
	"sync"
	"time"
)

type UI interface {
	Output(format string, a ...interface{})
	Message(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warning(format string, a ...interface{})
	Error(format string, a ...interface{})
}

const (
	timeFormat = "2006/01/02 15:04:05"
)

type ColoredUI struct {
	l sync.Mutex
}

func (c *ColoredUI) Output(format string, a ...interface{}) {
	c.l.Lock()
	defer c.l.Unlock()
	color.White(format, a...)
}

func (ColoredUI) appendTime(stamp time.Time, str string) string {
	return fmt.Sprintf("%s [bitbucket pipeline] %s", stamp.Format(timeFormat), str)
}

func (c *ColoredUI) Message(format string, a ...interface{}) {
	now := time.Now()
	c.l.Lock()
	defer c.l.Unlock()
	color.White(c.appendTime(now, format), a...)
}

func (c *ColoredUI) Warning(format string, a ...interface{}) {
	now := time.Now()
	c.l.Lock()
	defer c.l.Unlock()
	color.Yellow(c.appendTime(now, format), a...)
}

func (c *ColoredUI) Error(format string, a ...interface{}) {
	now := time.Now()
	c.l.Lock()
	defer c.l.Unlock()
	color.Red(c.appendTime(now, format), a...)
}

func (c *ColoredUI) Info(format string, a ...interface{}) {
	now := time.Now()
	c.l.Lock()
	defer c.l.Unlock()
	color.Cyan(c.appendTime(now, format), a...)
}

var ui UI = &ColoredUI{}
