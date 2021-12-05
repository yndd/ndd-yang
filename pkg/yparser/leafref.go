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

package yparser

import (
	"fmt"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/yndd/ndd-yang/pkg/leafref"
	"github.com/yndd/ndd-yang/pkg/yentry"
)

// ProcessLeafRef processes the leafref and returns
// if this is a leafref and if so the leafrefs local and remote path
// if the leafRef is local or external to the resource
func ProcessLeafRef(e *yang.Entry, resfullPath string, activeResPath *gnmi.Path) (*gnmi.Path, *gnmi.Path, bool) {
	switch GetTypeName(e) {
	default:
		switch GetTypeKind(e) {
		case "leafref":
			//fmt.Printf("LeafRef Entry: %#v \n", e)
			//fmt.Printf("LeafRef Name: %#v \n", e.Name)

			pathReference, found := getLeafRefPathRefernce(e.Node.Statement())
			if !found {
				fmt.Printf("ERROR LEAFREF NOT FOUND: %v \n", e.Node.Statement())
			}
			fmt.Printf("LeafRef pathReference: %v \n", pathReference)
			splitData := strings.Split(pathReference, "\n")
			var path string
			var elem string
			var k string
			for _, s := range splitData {
				// strip the junk from the leafref to get a plain xpath
				//fmt.Printf("LeafRef Path: %s\n", s)
				s = strings.ReplaceAll(s, ";", "")
				s = strings.ReplaceAll(s, "\"", "")
				s = strings.ReplaceAll(s, " ", "")
				s = strings.ReplaceAll(s, "\t", "")
				fmt.Printf("LeafRef pathReference clean: %s\n", s)

				// split the leafref per "/" and split the element and key from the path
				// last element is the key
				// 2nd last element is the element
				split2data := strings.Split(s, "/")
				//fmt.Printf("leafRef Len Split2 %d\n", len(split2data))

				for i, s2 := range split2data {
					switch i {
					case 0: // the first element in the leafref split is typically "", since the string before the "/" is empty
						if s2 != "" { // if not empty ensure we use the right data and split the string before ":" sign
							path += "/" + strings.Split(s2, ":")[len(strings.Split(s2, ":"))-1]

						}
					case (len(split2data) - 1): // last element is the key
						k = strings.Split(s2, ":")[len(strings.Split(s2, ":"))-1]
					case (len(split2data) - 2): // 2nd last element is the element
						elem = strings.Split(s2, ":")[len(strings.Split(s2, ":"))-1]
					default: // any other element gets added to the list
						path += "/" + strings.Split(s2, ":")[len(strings.Split(s2, ":"))-1]

					}
				}
				// if no path element exits we take the root "/" path
				if path == "" {
					path = "/"
				}
				// if the path contains /.. this is a relative leafref path
				relativeIndex := strings.Count(path, "/..")
				if relativeIndex > 0 {
					// fmt.Printf("leafRef Relative Path: %s, Element: %s, Key: %s, '/..' count %d\n", path, elem, k, relativeIndex)
					// check if the final p contains relative indirection to the resourcePath -> "/.."
					resSplitData := strings.Split(RemoveFirstEntryFromXpath(resfullPath), "/")
					//fmt.Printf("ResPath Split Length: %d data: %v\n", len(resSplitData), resSplitData)
					var addString string
					for i := 1; i <= (len(resSplitData) - 1 - strings.Count(path, "/..")); i++ {
						addString += "/" + resSplitData[i]
					}
					//fmt.Printf("leafRef Absolute Path Add string: %s\n", addString)
					path = addString + strings.ReplaceAll(path, "/..", "")
				}
				//fmt.Printf("leafRef Absolute Path: %s, Element: %v, Key: %s, '/..' count %d\n", path, e, k, relativeIndex)

			}
			fmt.Printf("LeafRef Path: %s, Elem: %s, Key: %s\n", path, elem, k)
			remotePath := Xpath2GnmiPath(path, 0)
			//remotePath = appendPathElem2GnmiPath(remotePath, elem, []string{k})
			remotePath = &gnmi.Path{Elem: append(remotePath.GetElem(), &gnmi.PathElem{Name: elem, Key: map[string]string{k: ""}})}

			// build a gnmi path and remove the first entry since the yang contains a duplicate path
			// localPath := Xpath2GnmiPath(resfullPath, 1)
			// the last element hould be a key in the previous element
			//localPath = TransformPathToLeafRefPath(localPath)

			if strings.Contains(GnmiPath2XPath(remotePath, false), GnmiPath2XPath(activeResPath, false)) {
				// if the remotePath and the active Path match exactly we classify this in the external leafref category
				// since we dont allow multiple elments of the same key in the same resource
				// E.g. interface ethernet-1/1 which reference a lag should be resolved to another interface in
				// another resource and hence this should be classified as an external leafref
				if GnmiPath2XPath(remotePath, false) != GnmiPath2XPath(activeResPath, false) {
					// this is a local leafref within the resource
					// make the localPath and remotePath relative to the resource
					//fmt.Printf("localPath: %v, remotePath %v, activePath %v\n", localPath, remotePath, activeResPath)
					////////localPath = transformGnmiPathAsRelative2Resource(localPath, activeResPath)
					remotePath = transformGnmiPathAsRelative2Resource(remotePath, activeResPath)
					//fmt.Printf("localPath: %v, remotePath %v\n", localPath, remotePath)
					return &gnmi.Path{Elem: []*gnmi.PathElem{{Name: e.Name}}}, remotePath, true
				}

			}
			// leafref is external to the resource
			//fmt.Printf("localPath: %v, remotePath %v, activePath %v\n", localPath, remotePath, activeResPath)
			// make the localPath relative to the resource
			/////// localPath = transformGnmiPathAsRelative2Resource(localPath, activeResPath)
			//fmt.Printf("localPath: %v, remotePath %v\n", localPath, remotePath)

			return &gnmi.Path{Elem: []*gnmi.PathElem{{Name: e.Name}}}, remotePath, false
		}
	}
	return nil, nil, false
}

func getLeafRefPathRefernce(s *yang.Statement) (string, bool) {
	if s.Kind() != "path" {
		for _, s := range s.SubStatements() {
			//fmt.Printf("statement: %#v\n", s)
			pr, found := getLeafRefPathRefernce(s)
			if found {
				return pr, true
			}
		}
		return "", false
	} else {
		//fmt.Printf("statement: %#v\n", s)
		return s.NName(), true
	}

}

func ValidateLeafRef(rootPath *gnmi.Path, x1, x2 interface{}, definedLeafRefs []*leafref.LeafRef, rs *yentry.Entry) (bool, []*leafref.ResolvedLeafRef, error) {
	// a global indication if the leafRef resolution was successfull or not
	// we are positive so we initialize to true
	success := true
	// variable to return a fine grane result
	resultValidations := make([]*leafref.ResolvedLeafRef, 0)

	// for all defined leafrefs check if the local leafref exists
	// if the local leafref is resolved, validate if the remote leafref is present
	// if not the resource cannot be configured
	for _, leafRef := range definedLeafRefs {
		resolvedLeafRefs := []*leafref.ResolvedLeafRef{
			{
				LeafRef: &leafref.LeafRef{
					LocalPath: &gnmi.Path{Elem: make([]*gnmi.PathElem, 0)},
				},
			},
		}
		// find the resolved local leafref objects that exists in the data
		// that is supplied in the function
		resolvedLeafRefs = rs.ResolveLocalLeafRefs(rootPath, leafRef.LocalPath, x1, resolvedLeafRefs, 0)

		// for all the resolved leafrefs validate if the remote leafref exists
		for _, resolvedLeafRef := range resolvedLeafRefs {
			fmt.Printf("resolvedLeafRef localPath: %s, resolved: %t, value: %v\n", GnmiPath2XPath(resolvedLeafRef.LeafRef.LocalPath, true), resolvedLeafRef.Resolved, resolvedLeafRef.Value)
			fmt.Printf("resolvedLeafRef remotePath: %s\n", GnmiPath2XPath(leafRef.RemotePath, true))
			// Validate if the leaf ref is resolved
			if resolvedLeafRef.Resolved {
				// validate if the leafref is local or external to the resource
				var found bool
				var remotePath *gnmi.Path
				// return if the remoteLeafRef is local to the data or remote
				if isRemoteLeafRefExternal(rootPath, leafRef.RemotePath) {
					// external leafref
					// the remote path keys are to be resolved, some will come from the rootpath
					// the rest comes from the leafref resolution
					// e.g. /topolopgy[name=y]/node[name=x]
					// the topology part will come from the rootpath, while the node part is coming from the rleafref resolution
					fmt.Printf("resolvedLeafRef external rootPath: %s\n", GnmiPath2XPath(rootPath, true))
					remotePath = buildExternalRemotePath(rootPath, leafRef.RemotePath, resolvedLeafRef.Value)
					fmt.Printf("resolvedLeafRef external remotePath: %s\n", GnmiPath2XPath(remotePath, true))
					fmt.Printf("resolvedLeafRef external data: %s\n", x2)
					// find remote leafRef with rootpath: / and the global config data
					found = rs.IsPathPresent(&gnmi.Path{}, remotePath, resolvedLeafRef.Value, x2)
				} else {
					// local leafref
					remotePath = buildLocalRemotePath(rootPath, leafRef.RemotePath, resolvedLeafRef.Value)
					// find remote leafRef with original rootpath and the global config data
					found = rs.IsPathPresent(rootPath, remotePath, resolvedLeafRef.Value, x1)
				}
				if !found {
					success = false
				}
				resultValidation := &leafref.ResolvedLeafRef{
					LeafRef: &leafref.LeafRef{
						LocalPath:  resolvedLeafRef.LeafRef.LocalPath,
						RemotePath: remotePath,
					},
					Value:    resolvedLeafRef.Value,
					Resolved: found,
				}
				resultValidations = append(resultValidations, resultValidation)
			}
		}
	}
	return success, resultValidations, nil
}

func isRemoteLeafRefExternal(rootPath, remotePath *gnmi.Path) bool {
	if strings.Contains(GnmiPath2XPath(remotePath, false), GnmiPath2XPath(rootPath, false)) {
		// if the remotePath and the active Path match exactly we classify this in the external leafref category
		// since we dont allow multiple elments of the same key in the same resource
		// E.g. interface ethernet-1/1 which reference a lag should be resolved to another interface in
		// another resource and hence this should be classified as an external leafref
		if GnmiPath2XPath(remotePath, false) != GnmiPath2XPath(rootPath, false) {
			// remote leafref is local
			return false
		}
	}
	// remote leafref is external
	return true
}

// buildLocalRemotePath provides a relative path to the rootpath with the data filled out
func buildLocalRemotePath(rootPath, remotePath *gnmi.Path, value string) *gnmi.Path {
	// cut the rootpath to get a relative path to the resource as remotePath
	newPath := &gnmi.Path{Elem: remotePath.GetElem()[(len(rootPath.GetElem()) - 1):]}
	// add value to path
	return addValue2Path(newPath, value)
}

// buildRemotePath merges the overlapping elements of the rootpath with the remotePath
func buildExternalRemotePath(rootPath, remotePath *gnmi.Path, value string) *gnmi.Path {
	// 1. create a new path based on the keys from the rootpath + remove the common elements from the remotePath
	// 2. populate the value in the remaining remotePath
	// 3. append the newPath with the remotePath
	rp := DeepCopyGnmiPath(remotePath)
	newPath := &gnmi.Path{Elem: make([]*gnmi.PathElem, 0)}

	for _, pathElem := range rootPath.GetElem() {
		// for each overlapping pathElem, except for the last one copy the pathElem from rootpath
		if pathElem.GetName() == rp.GetElem()[0].GetName() {
			newPath.Elem = append(newPath.GetElem(), pathElem)
			// cut the elemt from the remote path
			rp.Elem = rp.GetElem()[1:]
			// stop if the remote path is equal to 1
			if len(rp.GetElem()) == 1 {
				break
			}
		} else {
			break
		}
	}
	rp = addValue2Path(rp, value)
	newPath.Elem = append(newPath.GetElem(), rp.GetElem()...)
	return newPath
}

func addValue2Path(p *gnmi.Path, value string) *gnmi.Path {
	// for interface.subinterface we have a special handling where the value is seperated by a ethernet-1/1.4
	// the part before the dot represents the interface value in the key and the 2nd part represents the subinterface index
	// not sure how generic this is
	split := strings.Split(value, ".")
	n := 0
	for _, pathElem := range p.GetElem() {
		if len(pathElem.GetKey()) != 0 {
			for k := range pathElem.GetKey() {
				pathElem.GetKey()[k] = split[n]
				n++
			}
		}
	}
	return p
}
