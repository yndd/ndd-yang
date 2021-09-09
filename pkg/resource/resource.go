/*
Copyright 2020 Wim Henderickx.

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

package resource

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stoewer/go-strcase"
	"github.com/yndd/ndd-yang/pkg/container"
	"github.com/yndd/ndd-yang/pkg/parser"
)

type Resource struct {
	parser             *parser.Parser                 // calls a library for parsing JSON/YANG elements
	Path               *gnmi.Path                     // relative path from the resource; the absolute path is assembled using the resurce hierarchy with dependsOn
	ActualPath         *gnmi.Path                     // ActualPath is a relative path from the resource with the actual key information; the absolute path is assembled using the resurce hierarchy with dependsOn
	DependsOn          *Resource                      // resource dependency
	Excludes           []*gnmi.Path                   // relative from the the resource
	FileName           string                         // the filename the resource is using to render out the config
	ResFile            *os.File                       // the file reference for writing the resource file
	RootContainerEntry *container.Entry               // this is the root element which is used to reference the hierarchical resource information
	Container          *container.Container           // root container of the resource
	LastContainerPtr   *container.Container           // pointer to the last container we process
	ContainerList      []*container.Container         // List of all containers within the resource
	ContainerLevel     int                            // the current container Level when processing the yang entries
	ContainerLevelKeys map[int][]*container.Container // the current container Level key list
	LocalLeafRefs      []*parser.LeafRefGnmi
	ExternalLeafRefs   []*parser.LeafRefGnmi
}

// Option can be used to manipulate Options.
type Option func(g *Resource)

func WithXPath(p string) Option {
	return func(r *Resource) {
		r.Path = r.parser.XpathToGnmiPath(p, 0)
	}
}

func WithDependsOn(d *Resource) Option {
	return func(r *Resource) {
		r.DependsOn = d
	}
}

func WithExclude(p string) Option {
	return func(r *Resource) {
		r.Excludes = append(r.Excludes, r.parser.XpathToGnmiPath(p, 0))
	}
}

func NewResource(opts ...Option) *Resource {
	r := &Resource{
		parser: parser.NewParser(),
		Path:   new(gnmi.Path),
		//DependsOn:          new(Resource),
		Excludes:           make([]*gnmi.Path, 0),
		RootContainerEntry: nil,
		Container:          nil,
		LastContainerPtr:   nil,
		ContainerList:      make([]*container.Container, 0),
		ContainerLevel:     0,
		ContainerLevelKeys: make(map[int][]*container.Container),
		LocalLeafRefs:      make([]*parser.LeafRefGnmi, 0),
		ExternalLeafRefs:   make([]*parser.LeafRefGnmi, 0),
	}

	for _, o := range opts {
		o(r)
	}

	r.ContainerLevelKeys[0] = make([]*container.Container, 0)

	return r
}

func (r *Resource) GetResourcePath() *gnmi.Path {
	return r.Path
}

func (r *Resource) AddLocalLeafRef(ll, rl *gnmi.Path) {
	// add key entries to local leafrefs
	for _, llpElem := range ll.GetElem() {
		for _, c := range r.ContainerList {
			//fmt.Printf(" Resource AddLocalLeafRef llpElem.GetName(): %s, ContainerName: %s\n", c.Name, llpElem.GetName())
			if c.Name == llpElem.GetName() {
				for _, e := range c.Entries {
					//fmt.Printf(" Resource AddLocalLeafRef llpElem.GetName(): %s, ContainerName: %s, ContainerEntryName: %s\n", c.Name, llpElem.GetName(), e.GetName())
					if e.GetName() == llpElem.GetName() {
						if e.GetKey() != "" {
							llpElem.Key = make(map[string]string)
							llpElem.Key[e.GetKey()] = ""
						}
					}
				}
			}
		}
	}
	r.LocalLeafRefs = append(r.LocalLeafRefs, &parser.LeafRefGnmi{
		LocalPath:  ll,
		RemotePath: rl,
	})
}

func (r *Resource) AddExternalLeafRef(ll, rl *gnmi.Path) {
	// add key entries to local leafrefs
	entries := make([]*container.Entry, 0)
	for i, llpElem := range ll.GetElem() {
		if i == 0 {
			for _, c := range r.ContainerList {
				//fmt.Printf(" Resource AddExternalLeafRef i: %d llpElem.GetName(): %s, ContainerName: %s\n", i, llpElem.GetName(), c.Name)
				if c.Name == llpElem.GetName() {
					entries = c.Entries
				}
			}
		}
		for _, e := range entries {
			//fmt.Printf(" Resource AddExternalLeafRef i: %d llpElem.GetName(): %s, EntryName: %s\n", i, llpElem.GetName(), e.GetName())
			if e.GetName() == llpElem.GetName() {
				if e.GetKey() != "" {
					llpElem.Key = make(map[string]string)
					llpElem.Key[e.GetKey()] = ""
				}
				if e.Next != nil {
					entries = e.Next.Entries
				}
			}
		}
	}
	r.ExternalLeafRefs = append(r.ExternalLeafRefs, &parser.LeafRefGnmi{
		LocalPath:  ll,
		RemotePath: rl,
	})
}

func (r *Resource) GetLocalLeafRef() []*parser.LeafRefGnmi {
	return r.LocalLeafRefs
}

func (r *Resource) GetExternalLeafRef() []*parser.LeafRefGnmi {
	return r.ExternalLeafRefs
}

func (r *Resource) GetResourceNameWithPrefix(prefix string) string {
	return strcase.UpperCamelCase(prefix + "-" + r.GetAbsoluteName())
}

func (r *Resource) AssignFileName(prefix, suffix string) {
	r.FileName = prefix + "-" + strcase.KebabCase(r.GetAbsoluteName()) + suffix
}

func (r *Resource) CreateFile(dir, subdir1, subdir2 string) (err error) {
	r.ResFile, err = os.Create(filepath.Join(dir, subdir1, subdir2, filepath.Base(r.FileName)))
	return err
}

func (r *Resource) CloseFile() error {
	return r.ResFile.Close()
}

func (r *Resource) ResourceLastElement() string {
	return r.Path.GetElem()[len(r.Path.GetElem())-1].GetName()
}

func (r *Resource) GetRelativeGnmiPath() *gnmi.Path {
	return r.Path
}

// root resource have a additional entry in the path which is inconsistent with hierarchical resources
// to provide consistencyw e introduced this method to provide a consistent result for paths
// used mainly for leafrefs for now
func (r *Resource) GetRelativeGnmiActualResourcePath() *gnmi.Path {
	if r.DependsOn != nil {
		return r.Path
	}
	actPath := *r.Path
	actPath.Elem = actPath.Elem[1:(len(actPath.GetElem()))]
	return &actPath
}

// GetPath returns the relative Path of the resource
// For the root resources we need to strip the first entry of the path since srl uses some prefix entry
func (r *Resource) GetPath() *gnmi.Path {
	if r.DependsOn != nil {
		return r.Path
	}
	// we need to remove the first entry of the PathElem of the root resource
	actPath := r.Path
	actPath.Elem = actPath.Elem[1:(len(actPath.GetElem()))]
	return actPath
}

func (r *Resource) GetRelativeXPath() *string {
	return r.parser.GnmiPathToXPath(r.Path, true)
}

func (r *Resource) GetAbsoluteName() string {
	e := findPathElemHierarchy(r)
	// trim the first element since nokia yang comes with a aprefix
	if len(e) > 1 {
		e = e[1:]
	}
	// we remove the "-" from the element names otherwise we get a name clash when we parse all the Yang information
	newElem := make([]*gnmi.PathElem, 0)
	for _, entry := range e {
		name := strings.ReplaceAll(entry.Name, "-", "")
		name = strings.ReplaceAll(name, "ethernetsegment", "esi")
		pathElem := &gnmi.PathElem{
			Name: name,
			Key:  entry.GetKey(),
		}
		newElem = append(newElem, pathElem)
	}
	//fmt.Printf("PathELem: %v\n", newElem)
	return r.parser.GnmiPathToName(&gnmi.Path{
		Elem: newElem,
	})
}

// root resource have a additional entry in the path which is inconsistent with hierarchical resources
// to provide consistency we introduced this method to provide a consistent result for paths
// used mainly for leafrefs for now
func (r *Resource) GetAbsoluteGnmiActualResourcePath() *gnmi.Path {
	actPath := &gnmi.Path{
		Elem: findPathElemHierarchy(r),
	}

	actPath.Elem = actPath.Elem[1:(len(actPath.GetElem()))]
	return actPath
}

func (r *Resource) GetAbsoluteGnmiPath() *gnmi.Path {
	return &gnmi.Path{
		Elem: findPathElemHierarchy(r),
	}
}

func (r *Resource) GetAbsoluteXPathWithoutKey() *string {
	return r.parser.GnmiPathToXPath(&gnmi.Path{
		Elem: findPathElemHierarchy(r),
	}, false)
}

func (r *Resource) GetAbsoluteXPath() *string {
	return r.parser.GnmiPathToXPath(&gnmi.Path{
		Elem: findPathElemHierarchy(r),
	}, true)
}

func (r *Resource) GetExcludeRelativeXPath() []string {
	e := make([]string, 0)
	for _, p := range r.Excludes {
		e = append(e, *r.parser.GnmiPathToXPath(p, true))
	}
	return e
}

func findPathElemHierarchy(r *Resource) []*gnmi.PathElem {
	if r.DependsOn != nil {
		fp := findPathElemHierarchy(r.DependsOn)
		fp = append(fp, r.Path.Elem...)
		return fp
	}
	return r.Path.GetElem()
}

func (r *Resource) GetRootContainerEntry() *container.Entry {
	return r.RootContainerEntry
}

func (r *Resource) SetRootContainerEntry(e *container.Entry) {
	r.RootContainerEntry = e
}

func (r *Resource) GetAbsoluteLevel() int {
	return len(r.GetAbsoluteGnmiActualResourcePath().GetElem())
}

func (r *Resource) GetHierarchicalElements() []*HeInfo {
	he := make([]*HeInfo, 0)
	if r.DependsOn != nil {
		he = findHierarchicalElements(r.DependsOn, he)
	}
	return he
}

func DeepCopyConfigPath(in *gnmi.Path) *gnmi.Path {
	out := &gnmi.Path{
		Elem: make([]*gnmi.PathElem, 0),
	}
	for _, elem := range in.Elem {
		pathElem := &gnmi.PathElem{
			Name: elem.Name,
		}
		if len(elem.Key) != 0 {
			pathElem.Key = make(map[string]string)
			for k, v := range elem.Key {
				pathElem.Key[k] = v
			}
		}
		out.Elem = append(out.Elem, pathElem)
	}
	return out
}

func AddPathElem(p *gnmi.Path, e *container.Entry) *gnmi.Path {
	elem := &gnmi.PathElem{}
	if e.Key == "" {

		elem.Name = e.GetName()
	} else {
		elem.Name = e.GetName()
		elem.Key = map[string]string{strings.Split(e.GetKey(), " ")[0]: ""}
	}
	p.Elem = append(p.Elem, elem)
	return p
}

func (r *Resource) GetInternalHierarchicalPaths() []*gnmi.Path {
	// paths collects all paths
	paths := make([]*gnmi.Path, 0)
	// allocate a new path
	path := &gnmi.Path{
		Elem: make([]*gnmi.PathElem, 0),
	}
	// add root container entry to path elem
	AddPathElem(path, r.RootContainerEntry)
	// append the path to the paths list
	paths = append(paths, path)

	for _, e := range r.ContainerList[0].Entries {
		if e.Next != nil {
			//fmt.Printf("GetInternalHierarchicalPaths Next Entry : %v, Container: %v", e, e.Next)
			paths = addInternalHierarchicalPath(paths, path, e)
		}
	}
	return paths
}

func addInternalHierarchicalPath(paths []*gnmi.Path, origPath *gnmi.Path, e *container.Entry) []*gnmi.Path {
	// copy the old path to a new path
	path := DeepCopyConfigPath(origPath)
	// add container entry to path elem
	AddPathElem(path, e)
	// append the path to the paths list
	paths = append(paths, path)
	for _, e := range e.Next.Entries {
		if e.Next != nil {
			//fmt.Printf("addInternalHierarchicalPath Next Entry : %v, Container: %v", e, e.Next)
			paths = addInternalHierarchicalPath(paths, path, e)
		}
	}
	return paths

}

func findHierarchicalElements(r *Resource, he []*HeInfo) []*HeInfo {
	h := &HeInfo{
		Name: r.RootContainerEntry.Name,
		Key:  r.RootContainerEntry.Key,
		Type: r.RootContainerEntry.Type,
	}
	he = append(he, h)
	if r.DependsOn != nil {
		he = findHierarchicalElements(r.DependsOn, he)
	}
	return he
}

type HeInfo struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
	Type string `json:"type,omitempty"`
}

func (r *Resource) GetActualGnmiFullPath() *gnmi.Path {
	actPath := &gnmi.Path{
		Elem: findActualPathElemHierarchy(r),
	}
	actPath.Elem = actPath.Elem[1:(len(actPath.GetElem()))]
	return actPath
}

func findActualPathElemHierarchy(r *Resource) []*gnmi.PathElem {
	if r.DependsOn != nil {
		fp := findActualPathElemHierarchy(r.DependsOn)
		pathElem := r.Path.GetElem()
		if r.RootContainerEntry.Key != "" {
			pathElem[len(r.Path.GetElem())-1].Key = make(map[string]string)
			pathElem[len(r.Path.GetElem())-1].Key[r.RootContainerEntry.Key] = r.RootContainerEntry.Type
		}
		fp = append(fp, pathElem...)
		return fp
	}
	pathElem := r.Path.GetElem()
	if r.RootContainerEntry.Key != "" {
		pathElem[len(r.Path.GetElem())-1].Key = make(map[string]string)
		pathElem[len(r.Path.GetElem())-1].Key[r.RootContainerEntry.Key] = r.RootContainerEntry.Type
	}
	return pathElem
}
