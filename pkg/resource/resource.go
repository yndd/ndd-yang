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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stoewer/go-strcase"
	"github.com/yndd/ndd-yang/pkg/container"
	"github.com/yndd/ndd-yang/pkg/parser"
	"github.com/yndd/ndd-yang/pkg/yparser"
)

type Resource struct {
	Module               string                         // Yang Module name of the resource
	parser               *parser.Parser                 // calls a library for parsing JSON/YANG elements
	Path                 *gnmi.Path                     // relative path from the resource; the absolute path is assembled using the resurce hierarchy with Parent
	ActualPath           *gnmi.Path                     // ActualPath is a relative path from the resource with the actual key information; the absolute path is assembled using the resurce hierarchy with Parent
	Parent               *Resource                      // resource dependency
	ParentPath           *gnmi.Path                     // the full path of the parent
	SubDepPath           []string                       // the subpath the resource is dependent on w/o the last element
	Children             []*Resource                    // the children of the resource
	Excludes             []*gnmi.Path                   // relative from the the resource
	FileName             string                         // the filename the resource is using to render out the config
	ResFile              *os.File                       // the file reference for writing the resource file
	RootContainerEntry   *container.Entry               // this is the root element which is used to reference the hierarchical resource information
	Container            *container.Container           // root container of the resource
	LastContainerPtr     *container.Container           // pointer to the last container we process
	ContainerList        []*container.Container         // List of all containers within the resource
	ContainerLevel       int                            // the current container Level when processing the yang entries
	ContainerLevelKeys   map[int][]*container.Container // the current container Level key list
	LocalLeafRefs        []*parser.LeafRefGnmi
	ExternalLeafRefs     []*parser.LeafRefGnmi
	HierResourceElements *HierResourceElements // this defines the hierarchical elements the resource is dependent upon (map[string]interface -> map[string]map[string]map[string]interface{})
	SubResources         []*gnmi.Path          // for the fine grain reosurce allocation we sometimes see we need subresources: e.g. ipam has rir and instance within the parent
}

// Option can be used to manipulate Options.
type Option func(g *Resource)

func WithXPath(p string) Option {
	return func(r *Resource) {
		r.Path = yparser.Xpath2GnmiPath(p, 0)
	}
}

/*
func WithParent(d *Resource) Option {
	return func(r *Resource) {
		r.Parent = d
	}
}
*/

func WithParentPath(p *gnmi.Path) Option {
	return func(r *Resource) {
		r.ParentPath = p
	}
}

func WithExclude(p string) Option {
	return func(r *Resource) {
		r.Excludes = append(r.Excludes, r.parser.XpathToGnmiPath(p, 0))
	}
}

func WithModule(m string) Option {
	return func(r *Resource) {
		r.Module = m
	}
}

func WithSubResources(s []*gnmi.Path) Option {
	return func(r *Resource) {
		r.SubResources = s
	}
}

func WithSubDepPath(s []string) Option {
	return func(r *Resource) {
		r.SubDepPath = s
	}
}

func NewResource(parent *Resource, opts ...Option) *Resource {
	r := &Resource{
		parser:               parser.NewParser(),
		Path:                 new(gnmi.Path),
		Parent:               parent,
		SubDepPath:           make([]string, 0),
		Excludes:             make([]*gnmi.Path, 0),
		RootContainerEntry:   nil,
		Container:            nil,
		LastContainerPtr:     nil,
		ContainerList:        make([]*container.Container, 0),
		ContainerLevel:       0,
		ContainerLevelKeys:   make(map[int][]*container.Container),
		LocalLeafRefs:        make([]*parser.LeafRefGnmi, 0),
		ExternalLeafRefs:     make([]*parser.LeafRefGnmi, 0),
		HierResourceElements: NewHierResourceElements(),
		Children:             make([]*Resource, 0),
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

func (r *Resource) GetModule() string {
	return r.Module
}

func (r *Resource) GetParent() *Resource {
	return r.Parent
}

func (r *Resource) GetParentPath() *gnmi.Path {
	return r.ParentPath
}

func (r *Resource) GetChildren() []*Resource {
	return r.Children
}

func (r *Resource) GetParentResource() string {
	if r.GetParent() != nil {
		return r.GetParent().GetAbsoluteName()
	}
	return ""
}

func (r *Resource) GetChildResources() []string {
	children := make([]string, 0)
	for _, child := range r.GetChildren() {
		children = append(children, child.GetAbsoluteName())
	}
	return children
}

func (r *Resource) AddChild(res *Resource) {
	r.Children = append(r.Children, res)
}

// GetHierResourceElement return the hierarchical resource element
func (r *Resource) GetHierResourceElement() *HierResourceElements {
	return r.HierResourceElements
}

func (r *Resource) GetSubResources() []*gnmi.Path {
	return r.SubResources
}

// GetActualSubResources contaians the full path of all subresources
// the first pathElement is a Dummy but is used in the structs so we retain it inn the response
func (r *Resource) GetActualSubResources() []*gnmi.Path {
	paths := make([]*gnmi.Path, 0)
	for _, subres := range r.SubResources {
		paths = append(paths, &gnmi.Path{
			Elem: findActualSubResourcePathElemHierarchyWithoutKeys(r, r.ParentPath, subres),
		})
	}
	return paths
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
						if len(e.GetKey()) != 0 {
							for _, key := range e.GetKey() {
								llpElem.Key = make(map[string]string)
								llpElem.Key[key] = ""
							}
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
				if len(e.GetKey()) != 0 {
					for _, key := range e.GetKey() {
						llpElem.Key = make(map[string]string)
						llpElem.Key[key] = ""
					}
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

func (r *Resource) GetHierResourceElements() *HierResourceElements {
	return r.HierResourceElements
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
	if len(r.Path.GetElem()) > 0 {
		return r.Path.GetElem()[len(r.Path.GetElem())-1].GetName()
	}
	return ""

}

func (r *Resource) GetRelativeGnmiPath() *gnmi.Path {
	return r.Path
}

// root resource have a additional entry in the path which is inconsistent with hierarchical resources
// to provide consistencyw e introduced this method to provide a consistent result for paths
// used mainly for leafrefs for now
func (r *Resource) GetRelativeGnmiActualResourcePath() *gnmi.Path {
	if r.Parent != nil {
		return r.Path
	}
	actPath := *r.Path
	actPath.Elem = actPath.Elem[1:(len(actPath.GetElem()))]
	return &actPath
}

// GetPath returns the relative Path of the resource
// For the root resources we need to strip the first entry of the path since srl uses some prefix entry
func (r *Resource) GetPath() *gnmi.Path {
	if r.Parent != nil {
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
		//name := strings.ReplaceAll(entry.Name, "-", "")
		//name = strings.ReplaceAll(name, "ethernetsegment", "esi")
		name := strings.ReplaceAll(entry.Name, "ethernetsegment", "esi")
		pathElem := &gnmi.PathElem{
			Name: name,
			Key:  entry.GetKey(),
		}
		newElem = append(newElem, pathElem)
	}
	//fmt.Printf("PathELem: %v\n", newElem)
	absoluteName := r.parser.GnmiPathToName(&gnmi.Path{
		Elem: newElem,
	})
	if absoluteName == "" {
		return "device"
	}
	return absoluteName
}

// root resource have an additional entry in the path which is inconsistent with hierarchical resources
// to provide consistency we introduced this method to provide a consistent result for paths
// used mainly for leafrefs for now
func (r *Resource) GetAbsoluteGnmiActualResourcePath() *gnmi.Path {
	actPath := &gnmi.Path{
		Elem: findActualPathElemHierarchyWithoutKeys(r, r.ParentPath),
	}
	if len(actPath.GetElem()) != 0 {
		actPath.Elem = actPath.Elem[1:(len(actPath.GetElem()))]
		return actPath
	}
	return &gnmi.Path{}

}

func (r *Resource) GetAbsoluteGnmiPath() *gnmi.Path {
	actPath := &gnmi.Path{
		Elem: findActualPathElemHierarchyWithoutKeys(r, r.ParentPath),
	}

	return actPath
}

func (r *Resource) GetAbsoluteXPathWithoutKey() *string {
	actPath := &gnmi.Path{
		Elem: findActualPathElemHierarchyWithoutKeys(r, r.ParentPath),
	}

	return r.parser.GnmiPathToXPath(actPath, false)
}

func (r *Resource) getActualPath() []*gnmi.PathElem {
	if r.Parent != nil {
		pathElem := r.Parent.getActualPath()
		fmt.Printf("fp1: %v\n", pathElem)
		pe := r.Path.GetElem()
		pathElem = append(pathElem, pe...)
		return pathElem
	}
	pathElem := r.Path.GetElem()
	fmt.Printf("getActualPath, pathElem: %v\n", pathElem)
	return pathElem
}

func (r *Resource) GetAbsoluteXPath() *string {
	actPath := &gnmi.Path{
		Elem: r.getActualPath(),
	}
	return r.parser.GnmiPathToXPath(actPath, true)

}

func (r *Resource) GetActualGnmiFullPathWithKeys() *gnmi.Path {
	actPath := &gnmi.Path{
		Elem: findActualPathElemHierarchyWithKeys(r, r.ParentPath),
	}
	// the first element is a dummy container we can skip
	if len(actPath.GetElem()) > 0 {
		actPath.Elem = actPath.Elem[1:(len(actPath.GetElem()))]
		return actPath
	}
	return &gnmi.Path{}

}

func (r *Resource) GetExcludeRelativeXPath() []string {
	e := make([]string, 0)
	for _, p := range r.Excludes {
		e = append(e, *r.parser.GnmiPathToXPath(p, true))
	}
	return e
}

func findPathElemHierarchy(r *Resource) []*gnmi.PathElem {
	if r.Parent != nil {
		fp := findPathElemHierarchy(r.Parent)
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
	if r.Parent != nil {
		he = findHierarchicalElements(r.Parent, he)
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
		elem.Key = make(map[string]string)
		for _, key := range e.GetKey() {
			elem.Key[key] = ""
		}
		fmt.Printf("AddPathElem Key: %v \n", elem.Key)
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
	if r.Parent != nil {
		he = findHierarchicalElements(r.Parent, he)
	}
	return he
}

type HeInfo struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
	Type string `json:"type,omitempty"`
}

// findActualSubResourcePathElemHierarchyWithoutKeys, first gooes to the root of the resource and trickles back
// to find the full resourcePath with all Path Elements but does not try to find the keys
// used before the generator is run or during the generator
func findActualSubResourcePathElemHierarchyWithoutKeys(r *Resource, dp *gnmi.Path, subp *gnmi.Path) []*gnmi.PathElem {
	if r.Parent != nil {
		// we first go to the root of the resource to find the path
		fp := findActualSubResourcePathElemHierarchyWithoutKeys(r.Parent, r.ParentPath, r.ParentPath)
		pathElem := subp.GetElem()
		fp = append(fp, pathElem...)
		return fp
	}
	pathElem := dp.GetElem()
	if len(dp.GetElem()) == 0 {
		pathElem = subp.GetElem()
	}
	return pathElem
}

// findActualPathElemHierarchyWithoutKeys, first gooes to the root of the resource and trickles back
// to find the full resourcePath with all Path Elements but does not try to find the keys
// used before the generator is run or during the generator
func findActualPathElemHierarchyWithoutKeys(r *Resource, dp *gnmi.Path) []*gnmi.PathElem {
	if r.Parent != nil {
		fmt.Printf("findActualPathElemHierarchyWithoutKeys: parentpath %s\n", yparser.GnmiPath2XPath(r.ParentPath, false))
		// we first go to the root of the resource to find the path
		fp := findActualPathElemHierarchyWithoutKeys(r.Parent, r.ParentPath)
		fmt.Printf("fp1: %v\n", fp)
		pathElem := r.Path.GetElem()
		fp = append(fp, pathElem...)
		fmt.Printf("fp2: %v\n", fp)
		return fp
	}
	pathElem := dp.GetElem()
	if len(dp.GetElem()) == 0 {
		pathElem = r.Path.GetElem()
	}
	fmt.Printf("findActualPathElemHierarchyWithoutKeys, pathElem: %v\n", pathElem)
	return pathElem
}

// findActualPathElemHierarchy, first gooes to the root of the resource and trickles back
// to find the full resourcePath with all Path Elements (Names, Keys)
// used after the generator is run, to get the full path including the keys of the pathElements
func findActualPathElemHierarchyWithKeys(r *Resource, dp *gnmi.Path) []*gnmi.PathElem {
	if r.Parent != nil {
		// we first go to the root of the resource to find the path
		fp := findActualPathElemHierarchyWithKeys(r.Parent, r.ParentPath)
		pathElem := getResourcePathElemWithKeys(r, r.Path)
		fp = append(fp, pathElem...)
		return fp
	}
	pathElem := getResourcePathElemWithKeys(r, dp)
	return pathElem
}

func getResourcePathElemWithKeys(r *Resource, dp *gnmi.Path) []*gnmi.PathElem {
	// align the path Element with the dependency Path
	nextContainer := &container.Container{}
	pathElem := dp.GetElem()
	// when we are at the root of the resource the dependency path is not present
	// we initialize with the resource Path
	if len(dp.GetElem()) == 0 {
		pathElem = r.Path.GetElem()
	}
	fmt.Printf("Path Elem: %v\n", pathElem)
	for i, pe := range pathElem {
		fmt.Printf("Index: %d, root Path length: %d length Path: %d\n", i, len(r.Path.GetElem()), len(pathElem))
		switch {
		case i == len(r.Path.GetElem())-1: // root of the resource
			//fmt.Printf("    Element at root of resource: %d, peName: %s, Key: %v \n",i, pe.GetName(), pe.Key)
			//fmt.Printf("       RootContainerEntry: %#v\n", r.RootContainerEntry)
			if r.RootContainerEntry != nil && r.RootContainerEntry.Key != "" {
				pe.Key = make(map[string]string)
				// multiple keys in yang are supplied as a string and they delineation is a space
				// we split them here so we have access to each key indivifually
				// we initialaize the type as string as a dummy type
				split := strings.Split(r.RootContainerEntry.Key, " ")
				for _, key := range split {
					if r.RootContainerEntry.Next != nil {
						pe.Key[key] = r.RootContainerEntry.Next.GetKeyType(key)
					}
				}
				pe.Key[r.RootContainerEntry.Key] = r.RootContainerEntry.Type
			}
			nextContainer = r.Container
		case i > len(r.Path.GetElem())-1:
			if nextContainer != nil {
				//fmt.Printf("       Container Entries: %#v\n", nextContainer.Entries)
				for _, ce := range nextContainer.Entries {
					fmt.Printf("    Element within resource: %d ceName: %s, peName: %s, Key: %v \n", i, ce.GetName(), pe.GetName(), ce.Key)
					if ce.Name == pe.GetName() {
						if ce.Key != "" {
							pe.Key = make(map[string]string)
							// multiple keys in yang are supplied as a string and they delineation is a space
							// we split them here so we have access to each key indivifually
							// we initialaize the type as string as a dymmy type
							split := strings.Split(ce.Key, " ")
							for _, key := range split {
								pe.Key[key] = ce.Next.GetKeyType(key)
							}
						}
						nextContainer = ce.Next
						break
					}
				}
			}
		}
		//fmt.Printf("  PathElem: %d Name: %s, Key: %v \n", i, pe.GetName(), pe.GetKey())
	}
	return pathElem
}
