/*
Copyright 2021 Yndd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	Parent           *Entry
	Children         map[string]*Entry
	ResourceBoundary bool
	LeafRefs         []*leafref.LeafRef
}

type EntryOption func(Entry)

func (e *Entry) WithLogging(log logging.Logger) {
	e.Log = log
}

/*

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
	ResolveLocalLeafRefs(p *gnmi.Path, lrp *gnmi.Path, x interface{}, rlrs []*leafref.ResolvedLeafRef, lridx int) []*leafref.ResolvedLeafRef
	IsRemoteLeafRefPresent(p *gnmi.Path, rp *gnmi.Path, x interface{}) bool
}


type HandleInitFunc func(parent Handler, opts ...HandlerOption) Handler
*/
type EntryInitFunc func(parent Entry, opts ...EntryOption) Entry

func (e *Entry) GetName() string {
	return e.Name
}

func (e *Entry) GetKey() []string {
	return e.Key
}

func (e *Entry) GetParent() *Entry {
	return e.Parent
}

func (e *Entry) GetChildren() map[string]*Entry {
	return e.Children
}

func (e *Entry) GetResourceBoundary() bool {
	return e.ResourceBoundary
}

func (e *Entry) GetLeafRef() []*leafref.LeafRef {
	return e.LeafRefs
}

// GetKeys return the list of keys
func (e *Entry) GetKeys(p *gnmi.Path) []string {
	if len(p.GetElem()) != 0 {
		return e.Children[p.GetElem()[0].GetName()].GetKeys(&gnmi.Path{Elem: p.GetElem()[1:]})
	} else {
		return e.GetKey()
	}
}

// GetHierarchicalResourcesRemote returns the hierarchical paths of a resource
// 1. p is the path of the root resource
// 2. cp is the current path that extends to find the hierarchical resources once p is found
// 3. hierPaths contains the hierarchical resources
func (e *Entry) GetHierarchicalResourcesRemote(p *gnmi.Path, cp *gnmi.Path, hierPaths []*gnmi.Path) []*gnmi.Path {
	if len(p.GetElem()) != 0 {
		// continue finding the root of the resource we want to get the data from
		hierPaths = e.Children[p.GetElem()[0].GetName()].GetHierarchicalResourcesRemote(&gnmi.Path{Elem: p.GetElem()[1:]}, cp, hierPaths)
	} else {
		// we execute on a remote resource otherwise you collect the local information
		for _, h := range e.Children {
			newcp := &gnmi.Path{Elem: append(cp.GetElem(), &gnmi.PathElem{Name: h.GetName()})}
			if h.GetResourceBoundary() {
				hierPaths = append(hierPaths, newcp)
			} else {
				hierPaths = h.GetHierarchicalResourcesRemote(p, newcp, hierPaths)
			}
		}
	}
	return hierPaths
}

// GetHierarchicalResourcesLocal returns the hierarchical paths of a resource
// 0. root is to know the first resource that is actually the root of the path
// 1. p is the path of the root resource
// 2. cp is the current path that extends to find the hierarchical resources once p is found
// 3. hierPaths contains the hierarchical resources
func (e *Entry) GetHierarchicalResourcesLocal(root bool, p *gnmi.Path, cp *gnmi.Path, hierPaths []*gnmi.Path) []*gnmi.Path {
	if len(p.GetElem()) != 0 {
		// continue finding the root of the resource we want to get the data from
		hierPaths = e.Children[p.GetElem()[0].GetName()].GetHierarchicalResourcesLocal(root, &gnmi.Path{Elem: p.GetElem()[1:]}, cp, hierPaths)
	} else {
		newcp := cp
		if !root {
			newcp = &gnmi.Path{Elem: append(cp.GetElem(), &gnmi.PathElem{Name: e.GetName()})}
			if e.ResourceBoundary {
				hierPaths = append(hierPaths, newcp)
				return hierPaths
			}
		}
		for _, h := range e.Children {
			hierPaths = h.GetHierarchicalResourcesLocal(false, p, newcp, hierPaths)
		}
	}
	return hierPaths
}
