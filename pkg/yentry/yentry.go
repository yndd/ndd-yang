package yentry

import (
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-runtime/pkg/logging"
)

type Entry struct {
	Name     string
	Key      []string
	Parent   Handler
	Children map[string]Handler
}

type HandlerOption func(Handler)

func WithLogging(log logging.Logger) HandlerOption {
	return func(o Handler) {
		o.WithLogging(log)
	}
}

type Handler interface {
	WithLogging(log logging.Logger)
	GetKeys(p *gnmi.Path) []string
}

type HandleInitFunc func(parent Handler) Handler

func (e *Entry) GetName() string {
	return e.Name
}

func (e *Entry) GetKey() []string {
	return e.Key
}

func (e *Entry) GetParent() Handler {
	return e.Parent
}

func (e *Entry) GetChildren() map[string]Handler {
	return e.Children
}
