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
	"sort"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
)

// Xpath2GnmiPath convertss a xpath string to a gnmi path
func Xpath2GnmiPath(xpath string, offset int) (path *gnmi.Path) {
	split := strings.Split(xpath, "/")
	for i, element := range split {
		// ignore the first element
		if i == 0 {
			path = &gnmi.Path{
				Elem: make([]*gnmi.PathElem, 0),
			}
		} else {
			// offset is used to ignore an element from the path
			if i > offset {
				pathElem := &gnmi.PathElem{}
				if element != "" {
					if strings.Contains(element, "[") {
						s1 := strings.Split(element, "[")
						pathElem.Key = make(map[string]string)
						pathElem.Name = s1[0]
						s2 := strings.Split(s1[1], ",")
						for _, eWithKey := range s2 {
							// TODO if there is a "/" in the name of the path
							s := strings.Split(eWithKey, "=")
							var v string
							if strings.Contains(s[1], "]") {
								v = strings.Trim(s[1], "]")
							} else {
								v = s[1]
							}
							// trim blanks from the final element/key/value
							pathElem.Key[strings.Trim(s[0], " ")] = strings.Trim(v, " ")
						}
					} else {
						// trim blanks from the final element/key/value
						pathElem.Name = strings.Trim(element, " ")
					}
					path.Elem = append(path.Elem, pathElem)
				}
			}
		}
	}
	return path
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

// RemoveFirstEntryFromXpath removes the first entry of the xpath,
// so it trims the first element of the /
func RemoveFirstEntryFromXpath(s string) string {
	split := strings.Split(s, "/")
	var path string
	for i, s := range split {
		if i > 1 {
			path += "/" + s
		}
	}
	return path
}

func DeepCopyGnmiPath(in *gnmi.Path) *gnmi.Path {
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

// AppendPathElem2GnmiPath adds a pathElem to the config gnmi path
func AppendPathElem2GnmiPath(path *gnmi.Path, name string, keys []string) *gnmi.Path {
	pathElem := &gnmi.PathElem{
		Name: name,
	}
	if len(keys) > 0 {
		pathElem.Key = make(map[string]string)
		for _, key := range keys {
			pathElem.Key[key] = ""
		}
	}

	path.Elem = append(path.Elem, pathElem)
	return path
}

// TransformPathAsRelative2Resource returns a relative path
func TransformGnmiPathAsRelative2Resource(localPath, activeResPath *gnmi.Path) *gnmi.Path {
	if len(activeResPath.GetElem()) >= 1 {
		localPath.Elem = localPath.Elem[(len(activeResPath.GetElem()) - 1):(len(localPath.GetElem()))]
	} else {
		localPath.Elem = localPath.Elem[:len(localPath.GetElem())]
	}

	return localPath
}
