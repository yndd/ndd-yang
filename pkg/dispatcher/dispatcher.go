package dispatcher

import (
	"github.com/openconfig/gnmi/path"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-yang/pkg/cache"
	"github.com/yndd/ndd-yang/pkg/dtree"
)

var Resources = map[string][]*EventHandler{}

func Register(name string, e []*EventHandler) {
	Resources[name] = e
}

// A EventHandlerKind represents a kind of event handler
type EventHandlerKind string

// Operations Kinds.
const (
	// create
	EventHandlerCreate EventHandlerKind = "Create"
	// update
	EventHandlerEvent EventHandlerKind = "Event"
)

func (o *EventHandlerKind) String() string {
	return string(*o)
}

type EventHandler struct {
	PathElem []*gnmi.PathElem
	Kind     EventHandlerKind
	Handler  HandleConfigEventFunc
}

type HandleConfigEventFunc func(log logging.Logger, cc, sc *cache.Cache, prefix *gnmi.Path, p []*gnmi.PathElem, d interface{}) Handler

type Dispatcher struct {
	t *dtree.Tree
}

type DispatcherData struct {
	Kind    EventHandlerKind
	Handler func(log logging.Logger, cc, sc *cache.Cache, prefix *gnmi.Path, p []*gnmi.PathElem, d interface{}) Handler
}

type DispatcherConfig struct {
	PathElem []*gnmi.PathElem
}

func New() *Dispatcher {
	return &Dispatcher{
		t: &dtree.Tree{},
	}
}

func (r *Dispatcher) Init() {
	for _, ehs := range Resources {
		for _, eh := range ehs {
			r.Register(eh.PathElem, DispatcherConfig{
				PathElem: eh.PathElem,
			})
		}
	}
}

func (r *Dispatcher) GetTree() *dtree.Tree {
	return r.t
}

func (r *Dispatcher) Register(pe []*gnmi.PathElem, d interface{}) error {
	pathString := path.ToStrings(&gnmi.Path{Elem: pe}, false)
	return r.GetTree().Add(pathString, d)
}

func (r *Dispatcher) GetHandlerFunc(path []string) interface{} {
	return r.GetTree().GetLeafValue(path)
}

func (r *Dispatcher) GetPathElem(p *gnmi.Path) []*gnmi.PathElem {
	pathString := path.ToStrings(p, false)
	x := r.GetTree().GetLpm(pathString)
	o, ok := x.(DispatcherConfig)
	if !ok {
		return nil
	}
	return o.PathElem
}

// A Operation represents a crud operation
type Operation string

// Operations Kinds.
const (
	// create
	//OperationCreate Operation = "Create"
	// update
	OperationUpdate Operation = "Update"
	// delete
	OperationDelete Operation = "Delete"
)

func (o *Operation) String() string {
	return string(*o)
}
