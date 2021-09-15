/*
Copyright 2021 Wim Henderickx.

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

package parser

import (
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/yndd/ndd-runtime/pkg/logging"
)

type LeafRefValidationKind string

const (
	LeafRefValidationLocal    LeafRefValidationKind = "local"
	LeafRefValidationExternal LeafRefValidationKind = "external"
)

type LeafRefGnmi struct {
	LocalPath  *gnmi.Path `json:"localPath,omitempty"`
	RemotePath *gnmi.Path `json:"remotePath,omitempty"`
}

type ResolvedLeafRefGnmi struct {
	LocalPath  *gnmi.Path `json:"localPath,omitempty"`
	RemotePath *gnmi.Path `json:"remotePath,omitempty"`
	Value      string     `json:"value,omitempty"`
	Resolved   bool       `json:"resolved,omitempty"`
}

func (p *Parser) DeepCopyResolvedLeafRefGnmi(in *ResolvedLeafRefGnmi) (out *ResolvedLeafRefGnmi) {
	out = new(ResolvedLeafRefGnmi)
	if in.LocalPath != nil {
		out.LocalPath = new(gnmi.Path)
		out.LocalPath.Elem = make([]*gnmi.PathElem, 0)
		for _, v := range in.LocalPath.GetElem() {
			elem := &gnmi.PathElem{}
			elem.Name = v.Name
			if len(v.GetKey()) != 0 {
				elem.Key = make(map[string]string)
				for key, value := range v.Key {
					elem.Key[key] = value
				}
			}
			out.LocalPath.Elem = append(out.LocalPath.Elem, elem)
		}
	}
	if in.RemotePath != nil {
		out.RemotePath = new(gnmi.Path)
		out.RemotePath.Elem = make([]*gnmi.PathElem, 0)
		for _, v := range in.RemotePath.GetElem() {
			elem := &gnmi.PathElem{}
			elem.Name = v.Name
			if len(v.GetKey()) != 0 {
				elem.Key = make(map[string]string)
				for key, value := range v.Key {
					elem.Key[key] = value
				}
			}
			out.RemotePath.Elem = append(out.RemotePath.Elem, elem)
		}
	}
	out.Resolved = in.Resolved
	out.Value = in.Value
	return out
}

// ProcessLeafRef processes the leafref and returns if a leafref localPath, remotePath and if the leafRef is local or external to the resource
// used for yang parser
func (p *Parser) ProcessLeafRefGnmi(e *yang.Entry, resfullPath string, activeResPath *gnmi.Path) (*gnmi.Path, *gnmi.Path, bool) {
	switch p.GetTypeName(e) {
	default:
		switch p.GetTypeKind(e) {
		case "leafref":
			//fmt.Println(e.Node.Statement().String())
			splitData := strings.Split(e.Node.Statement().NName(), "\n")
			var path string
			var elem string
			var k string
			for _, s := range splitData {
				if strings.Contains(s, "path ") {
					// strip the junk from the leafref to get a plain xpath
					//fmt.Printf("LeafRef Path: %s\n", s)
					s = strings.ReplaceAll(s, "path ", "")
					s = strings.ReplaceAll(s, ";", "")
					s = strings.ReplaceAll(s, "\"", "")
					s = strings.ReplaceAll(s, " ", "")
					s = strings.ReplaceAll(s, "\t", "")
					//fmt.Printf("LeafRef Path: %s\n", s)

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
						//fmt.Printf("leafRef Relative Path: %s, Element: %s, Key: %s, '/..' count %d\n", path, elem, k, relativeIndex)
						// check if the final p contains relative indirection to the resourcePath -> "/.."
						resSplitData := strings.Split(p.RemoveFirstEntry(resfullPath), "/")
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
			}
			//fmt.Printf("Path: %s, Elem: %s, Key: %s\n", path, elem, k)
			remotePath := p.XpathToGnmiPath(path, 0)
			p.AppendElemInGnmiPath(remotePath, elem, k)

			// build a gnmi path and remove the first entry since the yang contains a duplicate path
			localPath := p.XpathToGnmiPath(resfullPath, 1)
			// the last element hould be a key in the previous element
			//localPath = TransformPathToLeafRefPath(localPath)

			if strings.Contains(*p.GnmiPathToXPath(remotePath, false), *p.GnmiPathToXPath(activeResPath, false)) {
				// if the remotePath and the active Path match exactly we classify this in the external leafref category
				// since we dont allow multiple elments of the same key in the same resource
				// E.g. interface ethernet-1/1 which reference a lag should be resolve to another interface in
				// another resource and hence this should be classified as an external leafref
				if *p.GnmiPathToXPath(remotePath, false) != *p.GnmiPathToXPath(activeResPath, false) {
					// this is a local leafref within the resource
					// make the localPath and remotePath relative to the resource
					//fmt.Printf("localPath: %v, remotePath %v, activePath %v\n", localPath, remotePath, activeResPath)
					localPath = p.TransformGnmiPathAsRelative2Resource(localPath, activeResPath)
					remotePath = p.TransformGnmiPathAsRelative2Resource(remotePath, activeResPath)
					//fmt.Printf("localPath: %v, remotePath %v\n", localPath, remotePath)
					return localPath, remotePath, true
				}

			}
			// leafref is external to the resource
			//fmt.Printf("localPath: %v, remotePath %v, activePath %v\n", localPath, remotePath, activeResPath)
			// make the localPath relative to the resource
			localPath = p.TransformGnmiPathAsRelative2Resource(localPath, activeResPath)
			//fmt.Printf("localPath: %v, remotePath %v\n", localPath, remotePath)

			return localPath, remotePath, false
		}
	}
	return nil, nil, false
}

// ValidateLocalLeafRef validates the local leafred information on the local resource data
// first the local leafrefs are resolved and if they are resolved
// the remote leaf refs within the objects are located
/// based on the result this funciton return the result + information on the validation
func (p *Parser) ValidateLeafRefGnmi(kind LeafRefValidationKind, x1, x2 interface{}, definedLeafRefs []*LeafRefGnmi, log logging.Logger) (bool, []*ResolvedLeafRefGnmi, error) {

	// a global indication if the leafRef resolution was successfull or not
	// we are positive so we initialize to true
	success := true
	// we initialize a global list for finer information on the resolution
	resultResolvedLeafRefs := make([]*ResolvedLeafRefGnmi, 0)

	// for all defined leafrefs check if the local leafref exists
	// if the local leafref is resolved, validate if the remote leafref is present
	// if not the resource cannot be configured
	for _, leafRef := range definedLeafRefs {
		// Initialize 1 entry in resolvedLeafRefs, whether it will be resolved or not
		// will be indicated by the Resolved flag in the resolvedLeafRef
		resolvedLeafRefs := make([]*ResolvedLeafRefGnmi, 0)
		resolvedLeafRef := &ResolvedLeafRefGnmi{
			LocalPath:  p.DeepCopyGnmiPath(leafRef.LocalPath), // used for path
			RemotePath: p.DeepCopyGnmiPath(leafRef.RemotePath),
			Value:      "",
			Resolved:   false,
		}
		resolvedLeafRefs = append(resolvedLeafRefs, resolvedLeafRef)
		// resolve the leafreference
		tc := &TraceCtxtGnmi{
			Path:             p.DeepCopyGnmiPath(leafRef.LocalPath), // used to walk through the object -> this data will not be filled in
			Idx:              0,
			Msg:              make([]string, 0),
			ResolvedLeafRefs: resolvedLeafRefs,
			Action:           ConfigResolveLeafRef,
		}
		p.ParseTreeWithActionGnmi(x1, tc, 0, 0)

		/*
			if len(tc.ResolvedLeafRefs) > 1 {
				fmt.Printf("ValidateLeafRef, localpath:%s, tc: %v\n", *p.ConfigGnmiPathToXPath(tc.Path, true), tc)
				for _, resolvedLeafRef := range tc.ResolvedLeafRefs {
					fmt.Printf("ValidateLeafRef resolvedLeafRef value     :%v\n", resolvedLeafRef.Value)
					fmt.Printf("ValidateLeafRef resolvedLeafRef resolved  :%v\n", resolvedLeafRef.Resolved)
					fmt.Printf("ValidateLeafRef resolvedLeafRef local path:%v\n", *p.ConfigGnmiPathToXPath(resolvedLeafRef.LocalPath, true))
					fmt.Printf("ValidateLeafRef resolvedLeafRef remotepath:%v\n", *p.ConfigGnmiPathToXPath(resolvedLeafRef.RemotePath, true))
				}
			}
		*/

		// for all the resolved leafrefs validate if the remote leafref exists
		for _, resolvedLeafRef := range tc.ResolvedLeafRefs {
			// Validate if the leaf ref is resolved
			if resolvedLeafRef.Resolved {
				// populate the remote leaf ref key
				p.PopulateRemoteLeafRefKeyGnmi(resolvedLeafRef)
				// find the Remote leafRef in the JSON data

				tc := &TraceCtxtGnmi{
					Path:   p.DeepCopyGnmiPath(resolvedLeafRef.RemotePath), // used to walk through the object -> this data will not be filled in
					Idx:    0,
					Msg:    make([]string, 0),
					Action: ConfigTreeActionFind,
				}
				if kind == LeafRefValidationLocal {
					// use the local data supplied in x1 for the remote leafref resolution
					p.ParseTreeWithActionGnmi(x1, tc, 0, 0)
				} else {
					// use the external data supplied in x2 for the remote leafref resolution
					p.ParseTreeWithActionGnmi(x2, tc, 0, 0)
				}

				// check if the remote leafref got resolved
				if !tc.Found {
					success = false
				}
				// fill out information which will be returned
				resultResolvedLeafRef := &ResolvedLeafRefGnmi{
					LocalPath:  resolvedLeafRef.LocalPath,
					RemotePath: resolvedLeafRef.RemotePath,
					Value:      resolvedLeafRef.Value,
					Resolved:   tc.Found,
				}
				resultResolvedLeafRefs = append(resultResolvedLeafRefs, resultResolvedLeafRef)
			}
		}
	}
	return success, resultResolvedLeafRefs, nil
}

// ValidateParentDependency validates the parent resource dependency
// based on the result this function returns the result + information on the validation
// we use a get here since we resolved the values of the keys alreay
func (p *Parser) ValidateParentDependency(x1 interface{}, definedParentDependencies []*LeafRefGnmi, log logging.Logger) (bool, []*ResolvedLeafRefGnmi, error) {
	// a global indication if the leafRef resolution was successfull or not
	// we are positive so we initialize to true
	success := true
	// we initialize a global list for finer information on the resolution
	resultleafRefValidation := make([]*ResolvedLeafRefGnmi, 0)
	// for all defined parent dependencies check if the remote leafref exists
	for _, depLeafRef := range definedParentDependencies {
		// find the last item with a key and resolve this, since the rest of the elments dont matter
		// and allows for more geenric code accross multiple implementations
		// srl can have additional elments that dont matter
		lastKeyElemIdx := 0
		for i, pathElem := range depLeafRef.RemotePath.GetElem() {
			if len(pathElem.GetKey()) != 0 {
				lastKeyElemIdx = i
			}
		}
		// lastKeyElemIdx is the last index
		depLeafRef.RemotePath.Elem = depLeafRef.RemotePath.GetElem()[:(lastKeyElemIdx + 1)]

		// get the Remote leafRef in the JSON data
		tc := &TraceCtxtGnmi{
			Path:   p.DeepCopyGnmiPath(depLeafRef.RemotePath), // used to walk through the object -> this data will not be filled in
			Idx:    0,
			Msg:    make([]string, 0),
			Action: ConfigTreeActionGet,
		}

		p.ParseTreeWithActionGnmi(x1, tc, 0, 0)

		// check if the remote leafref got resolved
		if !tc.Found {
			success = false
		}
		// fill out information which will be returned
		resolvedLeafRefValidationResult := &ResolvedLeafRefGnmi{
			RemotePath: depLeafRef.RemotePath,
			Resolved:   tc.Found,
		}
		resultleafRefValidation = append(resultleafRefValidation, resolvedLeafRefValidationResult)

	}
	return success, resultleafRefValidation, nil
}

// NOT SURE IF A SINGLE VALUE IS SOMETHING THAT WILL BE OK ACCROSS THE BOARD
// ValidateParentDependency validates the parent resource dependency
// the remote leaf refs within the objects are located
/// based on the result this funciton return the result + information on the validation
func (p *Parser) ValidateParentDependencyGnmi(x1 interface{}, value string, definedParentDependencies []*LeafRefGnmi, log logging.Logger) (bool, []*ResolvedLeafRefGnmi, error) {
	// a global indication if the leafRef resolution was successfull or not
	// we are positive so we initialize to true
	success := true
	// we initialize a global list for finer information on the resolution
	resultleafRefValidation := make([]*ResolvedLeafRefGnmi, 0)

	// for all defined parent dependencies check if the remote leafref exists
	for _, leafRef := range definedParentDependencies {
		// A parent dependency can only have 1 resolved leafref
		// for structs/code reuse reasons we leverage the same structs
		resolvedLeafRef := &ResolvedLeafRefGnmi{
			RemotePath: p.DeepCopyGnmiPath(leafRef.RemotePath),
			Value:      value,
		}

		p.PopulateRemoteLeafRefKeyGnmi(resolvedLeafRef)
		// find the Remote leafRef in the JSON data
		tc := &TraceCtxtGnmi{
			Path:   p.DeepCopyGnmiPath(resolvedLeafRef.RemotePath), // used to walk through the object -> this data will not be filled in
			Idx:    0,
			Msg:    make([]string, 0),
			Value:  resolvedLeafRef.Value,
			Action: ConfigTreeActionFind,
		}

		p.ParseTreeWithActionGnmi(x1, tc, 0, 0)

		// check if the remote leafref got resolved
		if !tc.Found {
			success = false
		}
		// fill out information which will be returned
		resolvedLeafRefValidationResult := &ResolvedLeafRefGnmi{
			RemotePath: resolvedLeafRef.RemotePath,
			Value:      resolvedLeafRef.Value,
			Resolved:   tc.Found,
		}
		resultleafRefValidation = append(resultleafRefValidation, resolvedLeafRefValidationResult)

	}
	return success, resultleafRefValidation, nil
}
