package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshelling(t *testing.T) {
	data := `
image: python:2.7
pipelines:
 default:
   - step:
      script:
        - python --version
        - python myScript.py
`
	reader := strings.NewReader(data)
	result := ReadPipelineDef(reader)
	assert.Equal(t, "python:2.7", result.Image, "missing image")
	assert.Equal(t, "python --version", result.Pipelines.Default[0].Step.Scripts[0], "missing script")
}

func TestWithOptions(t *testing.T) {
	data := `
image: python:2.7
pipelines:
 default:
   - step:
      script:
        - python --version
        - python myScript.py
options:
 docker: true
`
	reader := strings.NewReader(data)
	result := ReadPipelineDef(reader)
	assert.Equal(t, true, result.Options.Docker, "docker option")
}
