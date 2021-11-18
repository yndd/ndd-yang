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
	"strconv"

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
		// check if the localPath is one of the keys in the path. If not add it to the leafref
		if len(cp.GetElem()) != 0 && len(cp.GetElem()[len(cp.GetElem())-1].GetKey()) != 0 {
			if _, ok := cp.GetElem()[len(cp.GetElem())-1].GetKey()[lr.LocalPath.GetElem()[0].GetName()]; ok {
				// don't add the localPath Elem to the leaf ref
				leafRefs = append(leafRefs, &leafref.LeafRef{
					LocalPath:  cp,
					RemotePath: lr.RemotePath,
				})
			} else {
				// the leaafref localPath Elem does not match any key
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

	}
	return leafRefs
}

// ResolveLeafRefs is a runtime function that resolves the leafrefs
// it recursively walks to the data and validates if the local leafref and data match up
// if the resolution is successfull the function returns the resolved leafrefs
func (e *Entry) ResolveLocalLeafRefs(p *gnmi.Path, lrp *gnmi.Path, x1 interface{}, rlrs []*leafref.ResolvedLeafRef, lridx int) {
	if len(p.GetElem()) != 0 {
		// continue finding the root of the resource we want to get the data from
		e.Children[p.GetElem()[0].GetName()].ResolveLocalLeafRefs(&gnmi.Path{Elem: p.GetElem()[1:]}, lrp, x1, rlrs, lridx)
	} else {
		// check length is for protection
		if len(lrp.GetElem()) >= 1 {
			// append the leafref pathElem to the resolved leafref
			rlrs[lridx].LocalPath = &gnmi.Path{Elem: append(rlrs[lridx].LocalPath.GetElem(), lrp.GetElem()[0])}
			// validate if the data matches the leafref pathElem
			if x, ok := isDataPresent(lrp, x1, 0); ok {
				// data element exists
				if len(lrp.GetElem()[0].GetKey()) != 0 {
					// when a key is present, we process a list which can have multiple entries that need to be resolved
					e.resolveLeafRefsWithKey(p, lrp, x, rlrs, lridx)
				} else {
					// data element exists without keys
					if len(lrp.GetElem()) == 1 {
						// we are at the end of the leafref
						// use the value of the initial data validation for the resolved value
						if value, ok := getStringValue(x); ok {
							rlrs[lridx].Value = value
							rlrs[lridx].Resolved = true
						}
					} else {
						// continue; remove the pathElem from the leafref
						e.Children[lrp.GetElem()[0].GetName()].ResolveLocalLeafRefs(p, &gnmi.Path{Elem: lrp.GetElem()[1:]}, x, rlrs, lridx)
					}
				}
			}
			// resolution failed
		}
	}
}

func isDataPresent(lrp *gnmi.Path, x interface{}, idx int) (interface{}, bool) {
	switch x1 := x.(type) {
	case map[string]interface{}:
		if x2, ok := x1[lrp.GetElem()[idx].GetName()]; ok {
			return x2, true
		}
	}
	return nil, false
}

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

func (e *Entry) resolveLeafRefsWithKey(p *gnmi.Path, lrp *gnmi.Path, x interface{}, rlrs []*leafref.ResolvedLeafRef, lridx int) {
	// data element exists with keys
	switch x1 := x.(type) {
	case []interface{}:
		// copy the remote learef in case we see multiple elements in a container list
		rlrOrig := rlrs[lridx].DeepCopy()
		for n, x2 := range x1 {
			switch x3 := x2.(type) {
			case map[string]interface{}:
				if len(lrp.GetElem()) == 1 {
					if value, found := resolveKey(lrp, x3); found {
						rlrs[lridx].Value = value
						rlrs[lridx].Resolved = true
					}
				} else {
					if findKey(lrp, x3) {
						if n > 1 {
							rlrs = append(rlrs, rlrOrig)
						}
						if len(lrp.GetElem()) == 2 {
							// end of the leafref with leaf
							rlrs[lridx].LocalPath = &gnmi.Path{Elem: append(rlrs[lridx].LocalPath.GetElem(), lrp.GetElem()[1])}
							if x, ok := isDataPresent(lrp, x2, 1); ok {
								if value, ok := getStringValue(x); ok {
									rlrs[lridx].Value = value
									rlrs[lridx].Resolved = true
								}
								// data type nok
							}
							// resolution failed
						} else {
							// continue; remove the pathElem from the leafref
							e.Children[lrp.GetElem()[1].GetName()].ResolveLocalLeafRefs(p, &gnmi.Path{Elem: lrp.GetElem()[1:]}, x2, rlrs, lridx)
						}
					}
				}
			default:
				// resolution failed
			}
		}
	default:
		// resolution failed
	}
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

func findKey(p *gnmi.Path, x map[string]interface{}) bool {
	for keyName, keyValue := range p.GetElem()[0].GetKey() {
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
	if len(p.GetElem()) != 0 {
		// continue finding the root of the resource we want to get the data from
		return e.Children[p.GetElem()[0].GetName()].IsPathPresent(&gnmi.Path{Elem: p.GetElem()[1:]}, rp, value, x1)
	} else {
		// check length is for protection
		if len(rp.GetElem()) >= 1 {
			if x, ok := isDataPresent(rp, x1, 0); ok {
				// data element exists
				if len(rp.GetElem()[0].GetKey()) != 0 {
					// when a key is present, check if one entry matches
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
										return e.Children[rp.GetElem()[0].GetName()].IsPathPresent(p, &gnmi.Path{Elem: rp.GetElem()[1:]}, value, v)
									}
								}
								// even if not found there might be other elements in the list that match
							}
						}
					}
				} else {
					// data element exists without keys
					if len(rp.GetElem()) == 1 {
						// check if the value matches, if so remote leafRef was found
						if value == "" {
							return true
						}
						return x == value
					} else {
						return e.Children[rp.GetElem()[0].GetName()].IsPathPresent(p, &gnmi.Path{Elem: rp.GetElem()[1:]}, value, x)
					}
				}
			}
		}
		return false
	}
}