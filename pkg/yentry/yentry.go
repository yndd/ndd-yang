package yentry

import (
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-yang/pkg/leafref"
)

type Entry struct {
	Log              logging.Logger
	Name             string
	Key              []string
	Parent           Handler
	Children         map[string]Handler
	ResourceBoundary bool
	LeafRefs         []*leafref.LeafRef
}

type HandlerOption func(Handler)

func WithLogging(log logging.Logger) HandlerOption {
	return func(o Handler) {
		o.WithLogging(log)
	}
}

type Handler interface {
	WithLogging(log logging.Logger)
	GetName() string
	GetKey() []string
	GetKeys(p *gnmi.Path) []string
	GetResourceBoundary() bool
	GetHierarchicalResourcesRemote(p *gnmi.Path, cp *gnmi.Path, hierPaths []*gnmi.Path) []*gnmi.Path
	GetHierarchicalResourcesLocal(root bool, p *gnmi.Path, cp *gnmi.Path, hierPaths []*gnmi.Path) []*gnmi.Path
	GetLeafRefsLocal(root bool, p *gnmi.Path, cp *gnmi.Path, leafRefs []*leafref.LeafRef) []*leafref.LeafRef
}

type HandleInitFunc func(parent Handler, opts ...HandlerOption) Handler

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

func (e *Entry) GetResourceBoundary() bool {
	return e.ResourceBoundary
}

func (e *Entry) GetLeafRef() []*leafref.LeafRef {
	return e.LeafRefs
}
