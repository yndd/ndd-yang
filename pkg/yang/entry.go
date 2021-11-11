package yang

type Entry struct {
	Name     string
	Key      []string
	Parent   Handler
	Children map[string]Handler
}

type Handler interface {
	
}

type HandleInitFunc func(parent interface{}) Handler