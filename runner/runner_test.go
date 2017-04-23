package runner

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestRun(t *testing.T) {
	var data = `
image: busybox
pipelines:
 default:
  - step:
     script:
       - ls
       - ps
`
	yml, err := ioutil.TempFile("", "test-yml")
	if err != nil {
		t.Fatal(err)
	}
	if e := ioutil.WriteFile(yml.Name(), []byte(data), 0600); e != nil {
		t.Fatal(e)
	}
	result := Run(yml.Name(), "")
	assert.Equal(t, 0, result)
}

func TestRunBadCommand(t *testing.T) {
	var data = `
image: busybox
pipelines:
 default:
  - step:
     script:
       - cat foo
`
	yml, err := ioutil.TempFile("", "test-yml")
	if err != nil {
		t.Fatal(err)
	}
	if e := ioutil.WriteFile(yml.Name(), []byte(data), 0600); e != nil {
		t.Fatal(e)
	}
	result := Run(yml.Name(), "")
	assert.Equal(t, 1, result)
}

func TestRunWithDocker(t *testing.T) {
	var data = `
image: busybox
pipelines:
 default:
  - step:
     script:
       - ls
       - ps
       - docker version
options:
 docker: true
`
	yml, err := ioutil.TempFile("", "test-yml")
	if err != nil {
		t.Fatal(err)
	}
	if e := ioutil.WriteFile(yml.Name(), []byte(data), 0600); e != nil {
		t.Fatal(e)
	}
	result := Run(yml.Name(), "")
	assert.Equal(t, 0, result)
}
