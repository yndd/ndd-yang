package dispatcher

import (
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
	WithLogging(log logging.Logger)
	WithStateCache(c *cache.Cache)
	WithConfigCache(c *cache.Cache)
	WithPrefix(p *gnmi.Path)
	WithPathElem(pe []*gnmi.PathElem)
	WithRootSchema(rs *yentry.Entry)
}

type HandlerOption func(Handler)

func WithLogging(log logging.Logger) HandlerOption {
	return func(o Handler) {
		o.WithLogging(log)
	}
}

// WithStateCache initializes the state cache.
func WithStateCache(c *cache.Cache) HandlerOption {
	return func(o Handler) {
		o.WithStateCache(c)
	}
}

// WithConfigCache initializes the config cache.
func WithConfigCache(c *cache.Cache) HandlerOption {
	return func(o Handler) {
		o.WithConfigCache(c)
	}
}

func WithPrefix(p *gnmi.Path) HandlerOption {
	return func(o Handler) {
		o.WithPrefix(p)
	}
}

func WithPathElem(pe []*gnmi.PathElem) HandlerOption {
	return func(o Handler) {
		o.WithPathElem(pe)
	}
}

func WithRootSchema(rs *yentry.Entry) HandlerOption {
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
