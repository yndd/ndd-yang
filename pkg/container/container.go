/*
Copyright 2020 Yndd.

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

package container

import (
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-yang/pkg/leafref"
)

type Container struct {
	Name            string             `json:"name,omitempty"`
	Entries         []*Entry           `json:"entries,omitempty"`
	Prev            *Container         `json:"prev,omitempty"`
	ResourceBoundry bool               `json:"resourceBoundry,omitempty"`
	LeafRefs        []*leafref.LeafRef `json:"leafRefs,omitempty"`
}

type ContainerOption func(c *Container)

func NewContainer(n string, resourceBoundry bool, prev *Container, opts ...ContainerOption) *Container {
	e := &Container{
		Name:            n,
		Entries:         make([]*Entry, 0),
		Prev:            prev,
		ResourceBoundry: resourceBoundry,
	}

	for _, o := range opts {
		o(e)
	}

	return e
}

func (c *Container) GetName() string {
	return c.Name
}

func (c *Container) GetEntries() []*Entry {
	return c.Entries
}

func (c *Container) GetKeyType(name string) string {
	if c.Entries != nil {
		for _, e := range c.GetEntries() {
			if e.Name == name {
				return e.Type
			}
		}
	}
	return "string"
}

func (c *Container) GetKeyNames() []string {
	n := make([]string, 0)
	if c.Entries != nil {
		for _, e := range c.GetEntries() {
			if e.GetKeyBool() {
				n = append(n, e.Name)
			}
		}
	}
	return n
}

func (c *Container) GetChildren() []string {
	n := make([]string, 0)
	if c.Entries != nil {
		for _, e := range c.GetEntries() {
			if e.Next != nil {
				n = append(n, e.GetName())
			}
		}
	}
	return n
}

func (c *Container) GetSlicedFullName() []string {
	if c.Prev != nil {
		s := getRecursiveSlicedName(c.Prev)
		s = append(s, c.Name)
		return s
	}
	return []string{c.Name}
}

func getRecursiveSlicedName(c *Container) []string {
	if c.Prev != nil {
		s := getRecursiveSlicedName(c.Prev)
		s = append(s, c.Name)
		return s
	}
	return []string{c.Name}
}

// GetFullName replaces the dashes from the individual names to avoid clashes in names
// e.g, protocol bgp-evpn clashes with protocol bgp evpn
func (c *Container) GetFullName() string {
	if c.Prev != nil {
		return getRecursiveName(c.Prev) + "-" + strings.ReplaceAll(c.Name, "-", "")
	}
	return strings.ReplaceAll(c.Name, "-", "")
}
//replaces the dashes from the individual names to avoid clashes in names
// e.g, protocol bgp-evpn clashes with protocol bgp evpn
func getRecursiveName(c *Container) string {
	if c.Prev != nil {
		return getRecursiveName(c.Prev) + "-" + strings.ReplaceAll(c.Name, "-", "")
	}
	return strings.ReplaceAll(c.Name, "-", "")
}

func (c *Container) GetFullNameWithRoot() string {
	if c.Prev != nil {
		if getRecursiveNameWithRoot(c.Prev) == "" {
			return c.Name
		} else {
			return getRecursiveNameWithRoot(c.Prev) + "-" + c.Name
		}
	}
	return "root"
}

func getRecursiveNameWithRoot(c *Container) string {
	if c.Prev != nil {
		if getRecursiveNameWithRoot(c.Prev) == "" {
			return c.Name
		} else {
			return getRecursiveNameWithRoot(c.Prev) + "-" + c.Name
		}
	}
	return ""
}

func (c *Container) GetResourceBoundary() bool {
	return c.ResourceBoundry
}

func (c *Container) AddLeafRef(ll, rl *gnmi.Path) {
	c.LeafRefs = append(c.LeafRefs, &leafref.LeafRef{
		LocalPath:  ll,
		RemotePath: rl,
	})
}

func (c *Container) GetLeafRefs() []*leafref.LeafRef {
	return c.LeafRefs
}
