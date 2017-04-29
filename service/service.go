package service

const (
	PipelineServiceName       = "pipeline__service__"
)

type Service interface {
	Init()
	Start() ([]string, error)
	Attach(container string) error
	Stop()
}
