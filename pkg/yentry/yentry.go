package yentry

type Entry struct {
	Name     string
	Key      []string
	Parent   Handler
	Children map[string]Handler
}

type Handler interface {
	
}

type HandleInitFunc func(parent interface{}) Handler

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