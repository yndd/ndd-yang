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
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-yang/pkg/leafref"
)

// Returns all leafRefs for a given resource
// 1. p is the path of the root resource
// 2. cp is the current path that extends to find the hierarchical resources once p is found
// 3. leafRefs contains the leafRefs of the resource
func (e *Entry) GetLeafRefsLocal(root bool, p *gnmi.Path, cp *gnmi.Path, leafRefs []*leafref.LeafRef) []*leafref.LeafRef {
	if len(p.GetElem()) != 0 {
		// continue finding the root of the resource we want to get the data from
		leafRefs = e.Children[p.GetElem()[0].GetName()].GetLeafRefsLocal(root, &gnmi.Path{Elem: p.GetElem()[1:]}, cp, leafRefs)
	} else {
		newcp := deepCopyGnmiPath(cp)
		if !root {
			newcp = e.getNewPathWithKeys(cp)
			if e.ResourceBoundary {
				// when we are at the boundary we can return, since the leafref does not belong to this resource
				return leafRefs
			} else {
				leafRefs = e.appendLeafRefs(newcp, leafRefs)
			}
		} else {
			// append leafrefs of the root resource
			leafRefs = e.appendLeafRefs(newcp, leafRefs)
		}
		for _, h := range e.Children {
			leafRefs = h.GetLeafRefsLocal(false, p, newcp, leafRefs)
		}
	}
	return leafRefs
}

// getNewPathWithKeys return a new path with or without keys
func (e *Entry) getNewPathWithKeys(cp *gnmi.Path) *gnmi.Path {
	if len(e.GetKey()) != 0 {
		keys := make(map[string]string)
		for _, key := range e.GetKey() {
			keys[key] = ""
		}
		// return path with keys
		return &gnmi.Path{Elem: append(cp.GetElem(), &gnmi.PathElem{Name: e.GetName(), Key: keys})}
	}
	// return path without keys
	return &gnmi.Path{Elem: append(cp.GetElem(), &gnmi.PathElem{Name: e.GetName()})}
}

// appendLeafRefs extends the leafref path
func (e *Entry) appendLeafRefs(cp *gnmi.Path, leafRefs []*leafref.LeafRef) []*leafref.LeafRef {
	for _, lr := range e.GetLeafRef() {
		/*
			// check if the localPath is one of the keys in the path. If not add it to the leafref
			if len(cp.GetElem()) != 0 && len(cp.GetElem()[len(cp.GetElem())-1].GetKey()) != 0 {
				if _, ok := cp.GetElem()[len(cp.GetElem())-1].GetKey()[lr.LocalPath.GetElem()[0].GetName()]; ok {
					// don't add the localPath Elem to the leaf ref
					leafRefs = append(leafRefs, &leafref.LeafRef{
						LocalPath:  cp,
						RemotePath: lr.RemotePath,
					})
				} else {
					// the leafref localPath Elem does not match any key
					// // -> add the localPath Elem to the leaf ref
					leafRefs = append(leafRefs, &leafref.LeafRef{
						LocalPath:  &gnmi.Path{Elem: append(cp.GetElem(), &gnmi.PathElem{Name: lr.LocalPath.GetElem()[0].GetName()})},
						RemotePath: lr.RemotePath,
					})
				}
			} else {
				// current path Elem does not exist and there is also no key in the current path
				// -> add the localPath Elem to the leaf ref
				leafRefs = append(leafRefs, &leafref.LeafRef{
					LocalPath:  &gnmi.Path{Elem: append(cp.GetElem(), &gnmi.PathElem{Name: lr.LocalPath.GetElem()[0].GetName()})},
					RemotePath: lr.RemotePath,
				})
			}
		*/
		leafRefs = append(leafRefs, &leafref.LeafRef{
			LocalPath:  &gnmi.Path{Elem: append(cp.GetElem(), &gnmi.PathElem{Name: lr.LocalPath.GetElem()[0].GetName()})},
			RemotePath: lr.RemotePath,
		})

	}
	return leafRefs
}

// ResolveLeafRefs is a runtime function that resolves the leafrefs
// it recursively walks to the data and validates if the local leafref and data match up
// if the resolution is successfull the function returns the resolved leafrefs
func (e *Entry) ResolveLocalLeafRefs(p *gnmi.Path, lrp *gnmi.Path, x1 interface{}, resolution *leafref.Resolution, lridx int) {
	if len(p.GetElem()) != 0 {
		// continue finding the root of the resource we want to get the data from
		e.Children[p.GetElem()[0].GetName()].ResolveLocalLeafRefs(&gnmi.Path{Elem: p.GetElem()[1:]}, lrp, x1, resolution, lridx)
	} else {
		fmt.Printf("ResolveLocalLeafRefs yentry: lridx: %d, path: %s, leafrefpath: %s\n", lridx, GnmiPath2XPath(p, true), GnmiPath2XPath(lrp, true))
		fmt.Printf("ResolveLocalLeafRefs yentry: data: %v\n", x1)
		// check length is for protection
		if len(lrp.GetElem()) >= 1 {
			// append the leafref pathElem to the resolved leafref
			resolution.ResolvedLeafRefs[lridx].LocalPath = &gnmi.Path{Elem: append(resolution.ResolvedLeafRefs[lridx].LocalPath.GetElem(), lrp.GetElem()[0])}
			// validate if the data matches the leafref pathElem
			if x, ok := isDataPresent(lrp, x1, 0); ok {
				// data element exists
				if len(lrp.GetElem()[0].GetKey()) != 0 {
					// when a key is present, we process a list which can have multiple entries that need to be resolved
					e.resolveLeafRefsWithKey(p, lrp, x, resolution, lridx)
					fmt.Printf("ResolveLocalLeafRefs yentry: rlrs: %#v\n", resolution.ResolvedLeafRefs)
				} else {
					// data element exists without keys
					if len(lrp.GetElem()) == 1 {
						// we are at the end of the leafref
						// use the value of the initial data validation for the resolved value
						if value, ok := getStringValue(x); ok {
							resolution.ResolvedLeafRefs[lridx].Value = value
							resolution.ResolvedLeafRefs[lridx].Resolved = true
						}
					} else {
						// continue; remove the pathElem from the leafref
						fmt.Printf("entry name: %s\n", e.Name)
						for _, child := range e.Children {
							fmt.Printf("child name: %s\n", child.Name)
						}
						e.Children[lrp.GetElem()[0].GetName()].ResolveLocalLeafRefs(p, &gnmi.Path{Elem: lrp.GetElem()[1:]}, x, resolution, lridx)
					}
				}
			}
			// resolution failed
		}
	}
}

func isDataPresent(p *gnmi.Path, x interface{}, idx int) (interface{}, bool) {
	switch x1 := x.(type) {
	case map[string]interface{}:
		if x2, ok := x1[p.GetElem()[idx].GetName()]; ok {
			return x2, true
		}
	}
	return nil, false
}

/*
func resolveKey(p *gnmi.Path, x map[string]interface{}) (string, bool) {
	for keyName := range p.GetElem()[0].GetKey() {
		if x1, ok := x[keyName]; !ok {
			return "", false
		} else {
			if value, ok := getStringValue(x1); ok {
				return value, true
			} else {
				return "", false
			}
		}
	}
	return "", false
}
*/

func (e *Entry) resolveLeafRefsWithKey(p *gnmi.Path, lrp *gnmi.Path, x interface{}, resolution *leafref.Resolution, lridx int) {
	// data element exists with keys
	fmt.Printf("resolveLeafRefsWithKey1 yentry: lridx: %d, path: %s, leafrefpath: %s\n", lridx, GnmiPath2XPath(p, true), GnmiPath2XPath(lrp, true))
	fmt.Printf("resolveLeafRefsWithKey1 yentry: data: %v\n", x)
	switch x1 := x.(type) {
	case []interface{}:
		// copy the remote leafref in case we see multiple elements in a container list
		rlrOrig := resolution.ResolvedLeafRefs[lridx].DeepCopy()
		for n, x2 := range x1 {
			switch x3 := x2.(type) {
			case map[string]interface{}:
				fmt.Printf("resolveLeafRefsWithKey2 yentry n: %d, lrp: %s\n", n, GnmiPath2XPath(lrp, true))
				if n > 0 {
					resolution.ResolvedLeafRefs = append(resolution.ResolvedLeafRefs, rlrOrig)
					lridx++
				}
				insertKeyValueInLeafRef(lrp, x3, resolution, lridx)
				if len(lrp.GetElem()) == 2 {
					fmt.Printf("resolveLeafRefsWithKey3 yentry len=2, value: %v\n", x3[lrp.GetElem()[1].GetName()])
					// e.g. lrp will have endpoints[node-name=,interface-name=]/node-name
					if value, found := x3[lrp.GetElem()[1].GetName()]; found {
						if v, ok := getStringValue(value); ok {
							resolution.ResolvedLeafRefs[lridx].Value = v
							resolution.ResolvedLeafRefs[lridx].Resolved = true
						}
					}
				} else {
					e.Children[lrp.GetElem()[0].GetName()].ResolveLocalLeafRefs(p, &gnmi.Path{Elem: lrp.GetElem()[1:]}, x2, resolution, lridx)
					/*
						if findKey(lrp, x3) {
						fmt.Printf("key found\n")
						if len(lrp.GetElem()) == 2 {
							// end of the leafref with leaf
							resolution.ResolvedLeafRefs[lridx].LocalPath = &gnmi.Path{Elem: append(resolution.ResolvedLeafRefs[lridx].LocalPath.GetElem(), lrp.GetElem()[1])}
							if x, ok := isDataPresent(lrp, x2, 1); ok {
								if value, ok := getStringValue(x); ok {
									resolution.ResolvedLeafRefs[lridx].Value = value
									resolution.ResolvedLeafRefs[lridx].Resolved = true
								}
								// data type nok
							}
							// resolution failed
						} else {
							// continue; remove the pathElem from the leafref
							e.Children[lrp.GetElem()[1].GetName()].ResolveLocalLeafRefs(p, &gnmi.Path{Elem: lrp.GetElem()[1:]}, x2, resolution, lridx)
						}
					*/

				}
			default:
				// resolution failed
			}
		}
	default:
		// resolution failed
	}
	fmt.Printf("resolveLeafRefsWithKey4 yentry rlrs: %#v\n", resolution.ResolvedLeafRefs)

}

func getStringValue(x interface{}) (string, bool) {
	switch xx := x.(type) {
	case string:
		return string(xx), true
	case uint32:
		return strconv.Itoa(int(xx)), true
	case float64:
		return fmt.Sprintf("%.0f", xx), true
	default:
		return "", false
	}
}

func deepCopyGnmiPath(in *gnmi.Path) *gnmi.Path {
	out := new(gnmi.Path)
	if in != nil {
		out.Elem = make([]*gnmi.PathElem, 0)
		for _, pathElem := range in.GetElem() {
			elem := &gnmi.PathElem{
				Name: pathElem.GetName(),
			}
			if len(pathElem.GetKey()) != 0 {
				elem.Key = make(map[string]string)
				for keyName, keyValue := range pathElem.GetKey() {
					elem.Key[keyName] = keyValue
				}
			}
			out.Elem = append(out.Elem, elem)
		}
	}
	return out
}

func insertKeyValueInLeafRef(p *gnmi.Path, x map[string]interface{}, resolution *leafref.Resolution, lridx int) {
	// gather the keyValues from the data
	fmt.Printf("insertKeyValueInLeafRef: path: %s, key: %v, data: %v \n", GnmiPath2XPath(p, true), p.GetElem()[0].GetKey(), x)
	keys := make(map[string]string)
	for keyName := range p.GetElem()[0].GetKey() {
		if v, ok := x[keyName]; ok {
			switch x := v.(type) {
			case string:
				keys[keyName] = string(x)
			case uint32:
				keys[keyName] = strconv.Itoa(int(x)) 
			case float64:
				keys[keyName] = fmt.Sprintf("%.0f", x) 
			default:
				keys[keyName] = ""
			}
		}
	}
	path := deepCopyGnmiPath(resolution.ResolvedLeafRefs[lridx].LocalPath)
	for _, pe := range path.GetElem() {
		if pe.GetName() == p.GetElem()[0].GetName() {
			pe.Key = keys
		}
	}
	resolution.ResolvedLeafRefs[lridx].LocalPath = path
}

func findKey(p *gnmi.Path, x map[string]interface{}) bool {
	fmt.Printf("findKey: path %s, data: %v\n", GnmiPath2XPath(p, true), x)
	for keyName, keyValue := range p.GetElem()[0].GetKey() {
		fmt.Printf("findKey: keyName %s, keyValue: %s\n", keyName, keyValue)
		if v, ok := x[keyName]; !ok {
			return false
		} else {
			switch x := v.(type) {
			case string:
				if string(x) != keyValue {
					return false
				}
			case uint32:
				if strconv.Itoa(int(x)) != keyValue {
					return false
				}
			case float64:
				if fmt.Sprintf("%.0f", x) != keyValue {
					return false
				}
			default:
				return false
			}
		}
	}
	return true
}

func (e *Entry) IsPathPresent(p *gnmi.Path, rp *gnmi.Path, value string, x1 interface{}) bool {
	//fmt.Printf("IsPathPresent: rootpath: %s, remotePath: %s, value: %s\n", GnmiPath2XPath(p, true), GnmiPath2XPath(rp, true), value)
	//fmt.Printf("IsPathPresent: data: %v\n", x1)
	//fmt.Printf("IsPathPresent: len: %v\n", len(p.GetElem()))
	if len(p.GetElem()) != 0 {
		// continue finding the root of the resource we want to get the data from
		return e.Children[p.GetElem()[0].GetName()].IsPathPresent(&gnmi.Path{Elem: p.GetElem()[1:]}, rp, value, x1)
	} else {
		//fmt.Printf("IsPathPresent: rootpth: 0, remotePath: %s, value: %s\n", GnmiPath2XPath(rp, true), value)
		// check length is for protection
		if len(rp.GetElem()) >= 1 {
			pathElemName := rp.GetElem()[0].GetName()
			pathElemKey := rp.GetElem()[0].GetKey()
			if x, ok := isDataPresent(rp, x1, 0); ok {
				// data element exists
				if len(pathElemKey) != 0 {
					// when a key is present, check if one entry matches
					//fmt.Printf("IsPathPresent: data present with key remotePath: %s, data: %v\n", GnmiPath2XPath(rp, true), x)
					switch x1 := x.(type) {
					case []interface{}:
						for _, v := range x1 {
							switch x2 := v.(type) {
							case map[string]interface{}:
								if findKey(rp, x2) {
									if len(rp.GetElem()) == 1 {
										// remote leafref was found
										return true
									} else {
										return e.Children[pathElemName].IsPathPresent(p, &gnmi.Path{Elem: rp.GetElem()[1:]}, value, v)
									}
								}
								// even if not found there might be other elements in the list that match
							}
						}
					}
				} else {
					// data element exists without keys
					//fmt.Printf("IsPathPresent: data present without key remotePath: %s, data: %v\n", GnmiPath2XPath(rp, true), x)
					if len(rp.GetElem()) == 1 {
						// check if the value matches, if so remote leafRef was found
						if value == "" {
							return true
						}
						return x == value
					} else {
						return e.Children[pathElemName].IsPathPresent(p, &gnmi.Path{Elem: rp.GetElem()[1:]}, value, x)
					}
				}
			}
		}
		return false
	}
}

// GnmiPath2XPath converts a gnmi path with or without keys to a string pointer
func GnmiPath2XPath(path *gnmi.Path, keys bool) string {
	sb := strings.Builder{}
	for i, pElem := range path.GetElem() {
		pes := strings.Split(pElem.GetName(), ":")
		var pe string
		if len(pes) > 1 {
			pe = pes[1]
		} else {
			pe = pes[0]
		}
		sb.WriteString(pe)
		if keys {
			if len(pElem.GetKey()) != 0 {
				sb.WriteString("[")
				i := 0

				// we need to sort the keys in the same way for compaarisons
				type kv struct {
					Key   string
					Value string
				}
				var ss []kv
				for k, v := range pElem.GetKey() {
					ss = append(ss, kv{k, v})
				}
				sort.Slice(ss, func(i, j int) bool {
					return ss[i].Key > ss[j].Key
				})
				for _, kv := range ss {
					sb.WriteString(kv.Key)
					sb.WriteString("=")
					sb.WriteString(kv.Value)
					if i != len(ss)-1 {
						sb.WriteString(",")
					}
					i++
				}
				sb.WriteString("]")
			}
		}
		if i+1 != len(path.GetElem()) {
			sb.WriteString("/")
		}
	}
	return "/" + sb.String()
}
