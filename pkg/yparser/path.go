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
	"encoding/json"
	"sort"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-yang/pkg/leafref"
)

// GnmiPathToName converts a gnmi path to a name where each element of the
// path is seperated by a "-"
func GnmiPathToName(path *gnmi.Path) string {
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

		if i+1 != len(path.GetElem()) {
			sb.WriteString("-")
		}
	}
	return sb.String()
}

// Xpath2GnmiPath converts a xpath string to a gnmi path
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

/*
// AppendPathElem2GnmiPath adds a pathElem to the gnmi path
// used in leafref
func appendPathElem2GnmiPath(path *gnmi.Path, name string, keys []string) *gnmi.Path {
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
*/

// TransformPathAsRelative2Resource returns a relative path
func transformGnmiPathAsRelative2Resource(localPath, activeResPath *gnmi.Path) *gnmi.Path {
	if len(activeResPath.GetElem()) >= 1 {
		localPath.Elem = localPath.Elem[(len(activeResPath.GetElem()) - 1):(len(localPath.GetElem()))]
	} else {
		localPath.Elem = localPath.Elem[:len(localPath.GetElem())]
	}

	return localPath
}

// GetValue return the data of the gnmi typed value
func GetValue(updValue *gnmi.TypedValue) (interface{}, error) {
	if updValue == nil {
		return nil, nil
	}
	var value interface{}
	var jsondata []byte
	switch updValue.Value.(type) {
	case *gnmi.TypedValue_AsciiVal:
		value = updValue.GetAsciiVal()
	case *gnmi.TypedValue_BoolVal:
		value = updValue.GetBoolVal()
	case *gnmi.TypedValue_BytesVal:
		value = updValue.GetBytesVal()
	case *gnmi.TypedValue_DecimalVal:
		value = updValue.GetDecimalVal()
	case *gnmi.TypedValue_FloatVal:
		value = updValue.GetFloatVal()
	case *gnmi.TypedValue_IntVal:
		value = updValue.GetIntVal()
	case *gnmi.TypedValue_StringVal:
		value = updValue.GetStringVal()
	case *gnmi.TypedValue_UintVal:
		value = updValue.GetUintVal()
	case *gnmi.TypedValue_JsonIetfVal:
		jsondata = updValue.GetJsonIetfVal()
	case *gnmi.TypedValue_JsonVal:
		jsondata = updValue.GetJsonVal()
	case *gnmi.TypedValue_LeaflistVal:
		value = updValue.GetLeaflistVal()
	case *gnmi.TypedValue_ProtoBytes:
		value = updValue.GetProtoBytes()
	case *gnmi.TypedValue_AnyVal:
		value = updValue.GetAnyVal()
	}
	if value == nil && len(jsondata) != 0 {
		err := json.Unmarshal(jsondata, &value)
		if err != nil {
			return nil, err
		}
	}
	return value, nil
}

func GetRemotePathsFromResolvedLeafRef(resolvedLeafRef *leafref.ResolvedLeafRef) []*gnmi.Path {
	remotePaths := make([]*gnmi.Path, 0)
	for i := 0; i < len(strings.Split(resolvedLeafRef.Value, ".")); i++ {
		if i > 0 {
			// this is a special case where the value is split in "." e.g. network-instance -> interface + subinterface
			// or tunnel-interface + vxlan-interface
			// we create a shorter path to resolve the hierarchical path
			remotePath := &gnmi.Path{
				Elem: make([]*gnmi.PathElem, 0),
			}
			// we return on the first reference path
			for _, pathElem := range resolvedLeafRef.RemotePath.GetElem() {
				if len(pathElem.GetKey()) != 0 {
					remotePath.Elem = append(remotePath.Elem, &gnmi.PathElem{Name: pathElem.GetName(), Key: pathElem.GetKey()})
					// we stop at copying the first key
					break
				} else {
					remotePath.Elem = append(remotePath.Elem, &gnmi.PathElem{Name: pathElem.GetName()})
				}
			}
			remotePaths = append(remotePaths, remotePath)
		} else {
			remotePaths = append(remotePaths, resolvedLeafRef.RemotePath)
		}
	}
	return remotePaths
}

// GnmiPathToSubResourceName special case since we have to remove the - from the first elemen
// used in nddbuilder
func GnmiPathToSubResourceName(path *gnmi.Path) string {
	sb := strings.Builder{}
	for i, pElem := range path.GetElem() {
		pathElemName := pElem.GetName()
		if i == 0 {
			pathElemName = strings.ReplaceAll(pathElemName, "-", "")
		}
		pes := strings.Split(pathElemName, ":")
		var pe string
		if len(pes) > 1 {
			pe = pes[1]
		} else {
			pe = pes[0]
		}
		sb.WriteString(pe)

		if i+1 != len(path.GetElem()) {
			sb.WriteString("-")
		}
	}
	return sb.String()
}
