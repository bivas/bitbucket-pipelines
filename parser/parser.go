package parser

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type config struct {
	Step step `yaml:"step"`
}

type step struct {
	Scripts []string `yaml:"script"`
}

type pipelines struct {
	Default   []config `yaml:"default"`
	Branches  []config `yaml:"branches,omitempty"`
	Tags      []config `yaml:"tags,omitempty"`
	Bookmarks []config `yaml:"bookmarks,omitempty"`
}

type options struct {
	Docker bool `yaml:"docker,omitempty"`
}

//PipelineDef : bitbucket pipelines definition
type PipelineDef struct {
	Image     string    `yaml:"image"`
	Pipelines pipelines `yaml:"pipelines"`
	Options   options   `yaml:"options,omitempty"`
}

//ReadPipelineDef : parse bitbucket pipeline input
func ReadPipelineDef(reader io.Reader) (result PipelineDef) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(bytes, &result)
	if err != nil {
		panic(err)
	}
	return
}
