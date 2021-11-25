package dispatcher

import (
	"github.com/karimra/gnmic/types"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-yang/pkg/cache"
	"github.com/yndd/ndd-yang/pkg/yentry"
)

type Handler interface {
	HandleConfigEvent(o Operation, prefix *gnmi.Path, pe []*gnmi.PathElem, d interface{}) (Handler, error)
	SetParent(interface{}) error
	SetRootSchema(rs *yentry.Entry)
	GetChildren() map[string]string
	UpdateConfig(interface{}) error
	UpdateStateCache() error
	DeleteStateCache() error
	GetTargets() []*types.TargetConfig
	WithLogging(log logging.Logger)
	WithStateCache(c *cache.Cache)
	WithConfigCache(c *cache.Cache)
	WithPrefix(p *gnmi.Path)
	WithPathElem(pe []*gnmi.PathElem)
	WithRootSchema(rs *yentry.Entry)
}

type Option func(Handler)

func WithLogging(log logging.Logger) Option {
	return func(o Handler) {
		o.WithLogging(log)
	}
}

// WithStateCache initializes the state cache.
func WithStateCache(c *cache.Cache) Option {
	return func(o Handler) {
		o.WithStateCache(c)
	}
}

// WithConfigCache initializes the config cache.
func WithConfigCache(c *cache.Cache) Option {
	return func(o Handler) {
		o.WithConfigCache(c)
	}
}

func WithPrefix(p *gnmi.Path) Option {
	return func(o Handler) {
		o.WithPrefix(p)
	}
}

func WithPathElem(pe []*gnmi.PathElem) Option {
	return func(o Handler) {
		o.WithPathElem(pe)
	}
}

func WithRootSchema(rs *yentry.Entry) Option {
	return func(o Handler) {
		o.WithRootSchema(rs)
	}
}

type Resource struct {
	Log         logging.Logger
	ConfigCache *cache.Cache
	StateCache  *cache.Cache
	PathElem    *gnmi.PathElem
	Prefix      *gnmi.Path
	RootSchema  *yentry.Entry
	Key         string
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
