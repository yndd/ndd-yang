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
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/yndd/ndd-yang/pkg/leafref"
)

type Container struct {
	Name             string             `json:"name,omitempty"`
	Namespace        string             `json:"namespace,omitempty"`
	Prefix           *yang.Value        `json:"prefix,omitempty"`
	Module           string             `json:"module,omitempty"`
	ReadOnly         bool               `json:"read-only,omitempty"`
	HasState         bool               `json:"state-child,omitempty"`
	Entries          []*Entry           `json:"entries,omitempty"`
	Children         []*Container       `json:"children,omitempty"`
	Prev             *Container         `json:"prev,omitempty"`
	ResourceBoundary bool               `json:"resourceBoundry,omitempty"`
	LeafRefs         []*leafref.LeafRef `json:"leafRefs,omitempty"`
}

type ContainerOption func(c *Container)

func NewContainer(e *yang.Entry, namespace string, readOnly, resourceBoundary bool, prev *Container, opts ...ContainerOption) *Container {
	c := &Container{
		Name:             e.Name,
		Namespace:        namespace,
		Prefix:           e.Prefix,
		ReadOnly:         readOnly,
		Entries:          make([]*Entry, 0),
		Prev:             prev,
		Children:         make([]*Container, 0),
		ResourceBoundary: resourceBoundary,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *Container) GetName() string {
	return c.Name
}

func (c *Container) GetNamespace() string {
	return c.Namespace
}

func (c *Container) GetPrefixName() string {
	if c.Prefix == nil {
		return ""
	}
	return c.Prefix.Name
}

func (c *Container) GetModuleName() string {
	return c.Module
}

func (c *Container) GetReadOnly() bool {
	return c.ReadOnly
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

func (c *Container) GetChildren() []*Container {
	return c.Children
}

func (c *Container) GetChildrenNames() []string {
	n := make([]string, 0)
	for _, c := range c.GetChildren() {
		n = append(n, c.GetName())
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
			return strings.ReplaceAll(c.Name, "-", "")
		} else {
			return getRecursiveNameWithRoot(c.Prev) + "-" + strings.ReplaceAll(c.Name, "-", "")
		}
	}
	return "root"
}

func getRecursiveNameWithRoot(c *Container) string {
	if c.Prev != nil {
		if getRecursiveNameWithRoot(c.Prev) == "" {
			return strings.ReplaceAll(c.Name, "-", "")
		} else {
			return getRecursiveNameWithRoot(c.Prev) + "-" + strings.ReplaceAll(c.Name, "-", "")
		}
	}
	return ""
}

func (c *Container) GetResourceBoundary() bool {
	return c.ResourceBoundary
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

func (c *Container) SetHasState() {
	c.HasState = true
}

func (c *Container) AddContainerChild(cc *Container) {
	c.Children = append(c.Children, cc)
}

func (c *Container) UpdateHasState2ParentContainers() {
	if c.Prev != nil {
		c.Prev.SetHasState()
		c.Prev.UpdateHasState2ParentContainers()
	}
}

func (c *Container) SetModule(m string) {
	c.Module = m
}
