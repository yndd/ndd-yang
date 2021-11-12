package yentry

import "github.com/openconfig/gnmi/proto/gnmi"

type Entry struct {
	Name     string
	Key      []string
	Parent   Handler
	Children map[string]Handler
}

type Handler interface {
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
