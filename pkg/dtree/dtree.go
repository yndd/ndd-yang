package dtree

import (
	"sync"
)

type Tree struct {
	mu     sync.RWMutex
	branch map[string]*Branch
}

type Branch struct {
	mu    sync.RWMutex
	t     *Tree
	value interface{}
}

func (t *Branch) GetTree() *Tree {
	return t.t
}

func (t *Tree) GetBranch() map[string]*Branch {
	return t.branch
}

func (t *Tree) Add(path []string, value interface{}) error {
	switch len(path) {
	case 0:
		return nil
	case 1:
		return t.addTerminalValue(path, value)
	default:
		return t.intermediateAdd(path, value)
	}
}

func (t *Tree) addTerminalValue(path []string, value interface{}) error {
	defer t.mu.Unlock()
	t.mu.Lock()
	if len(t.branch) == 0 {
		t.branch = make(map[string]*Branch)
	}
	// if the element already exists dont override it
	if b, ok := t.branch[path[0]]; !ok {
		t.branch[path[0]] = &Branch{
			value: value,
			t:     &Tree{},
		}
	} else {
		b.value = value
	}
	return nil
}

func (t *Tree) intermediateAdd(path []string, value interface{}) error {
	defer t.mu.Unlock()
	t.mu.Lock()
	if len(t.branch) == 0 {
		t.branch = make(map[string]*Branch)
	}
	if _, ok := t.branch[path[0]]; !ok {
		t.branch[path[0]] = &Branch{
			t: &Tree{},
		}
	}

	return t.branch[path[0]].t.Add(path[1:], value)
}

// Get returns the Tree node if path points to it, nil otherwise.
// All nodes in path must be fully specified with no globbing (*).
func (t *Tree) Get(path []string) *Branch {
	defer t.mu.RUnlock()
	t.mu.RLock()

	switch len(path) {
	case 1:
		// check for a specific entry
		if b := t.branch[path[0]]; b != nil {
			// if the value is not initializedd we try to see if there is a * match
			if b.value == nil {
				if db := t.branch["*"]; db != nil {
					if db.value != nil {
						return db
					}
				}
			}
			return b
		}
		if b := t.branch["*"]; b != nil {
			return b
		}
		return nil
	default:
		// check for a specific entry
		if b := t.branch[path[0]]; b != nil {
			return b.t.Get(path[1:])
		}
		// if the above did not work check for a wildcard entry
		if b := t.branch["*"]; b != nil {
			return b.t.Get(path[1:])
		}
		return nil
	}
}

// GetLeafValue returns the leaf value if path points to a leaf in t, nil otherwise. All
// nodes in path must be fully specified with no globbing (*).
func (t *Tree) GetLeafValue(path []string) interface{} {
	return t.Get(path).Value()
}

// Value returns the latest value stored in node t if it represents a leaf, nil otherwise. Value is safe to call on
// nil Tree.
func (t *Branch) Value() interface{} {
	if t == nil {
		return nil
	}
	defer t.mu.RUnlock()
	t.mu.RLock()
	return t.value
}

func (t *Tree) GetLpm(path []string) interface{} {
	return t.getlpm(path).Value()
}

func (t *Tree) getlpm(path []string) *Branch {
	defer t.mu.RUnlock()
	t.mu.RLock()

	// check for a specific entry
	if b := t.branch[path[0]]; b != nil {
		// full match
		if len(path) == 1 {
			return b
		}
		if x := b.t.getlpm(path[1:]); x == nil {
			// the next value in the path was not found, so we return the current branch
			return b
		} else {
			return x
		}
	}
	// if the above did not work check for a wildcard entry
	if b := t.branch["*"]; b != nil {
		// full match
		if len(path) == 1 {
			return b
		}
		if x := b.t.getlpm(path[1:]); x == nil {
			// the next value in the path was not found, so we return the current branch
			return b
		} else {
			return x
		}
	}
	// not found
	return nil
}
