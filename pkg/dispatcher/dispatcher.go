package dispatcher

import (
	"github.com/openconfig/gnmi/path"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-yang/pkg/cache"
	"github.com/yndd/ndd-yang/pkg/dtree"
)

/*
var Resources = map[string][]*EventHandler{}

func Register(name string, e []*EventHandler) {
	Resources[name] = e
}

type EventHandler struct {
	PathElem []*gnmi.PathElem
	Kind     EventHandlerKind
	Handler  HandleConfigEventFunc
}
*/

/*
// A EventHandlerKind represents a kind of event handler
type EventHandlerKind string

// Operations Kinds.
const (
	// create
	EventHandlerCreate EventHandlerKind = "Create"
	// update
	//EventHandlerEvent EventHandlerKind = "Event"
)

func (o *EventHandlerKind) String() string {
	return string(*o)
}
*/

type HandleConfigEventFunc func(log logging.Logger, cc, sc *cache.Cache, prefix *gnmi.Path, p []*gnmi.PathElem, d interface{}) Handler

type Dispatcher interface{
	Init(resources []*gnmi.Path)
	GetTree() *dtree.Tree
	GetPathElem(p *gnmi.Path) []*gnmi.PathElem
}

type dispatcher struct {
	t *dtree.Tree
}

/*
type DispatcherData struct {
	//Kind    EventHandlerKind
	Handler HandleConfigEventFunc
}
*/

type dispatcherConfig struct {
	PathElem []*gnmi.PathElem
}

func New() Dispatcher {
	return &dispatcher{
		t: &dtree.Tree{},
	}
}

func (r *dispatcher) Init(resources []*gnmi.Path) {
	for _, path := range resources {
		r.register(path.GetElem(), dispatcherConfig{
			PathElem: path.GetElem(),
		})
	}
}

func (r *dispatcher) GetTree() *dtree.Tree {
	return r.t
}

func (r *dispatcher) register(pe []*gnmi.PathElem, d interface{}) error {
	pathString := path.ToStrings(&gnmi.Path{Elem: pe}, false)
	return r.GetTree().Add(pathString, d)
}

func (r *dispatcher) GetPathElem(p *gnmi.Path) []*gnmi.PathElem {
	pathString := path.ToStrings(p, false)
	x := r.GetTree().GetLpm(pathString)
	o, ok := x.(dispatcherConfig)
	if !ok {
		return nil
	}
	return o.PathElem
}