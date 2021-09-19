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
	"encoding/json"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-runtime/pkg/utils"
)

// GnmiPathToName converts a config gnmi path to a name where each element of the
// path is seperated by a "-"
func (p *Parser) GnmiPathToName(path *gnmi.Path) string {
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

// GnmiPathToXPath converts a gnmi path with or withour keys to a string pointer
func (p *Parser) GnmiPathToXPath(path *gnmi.Path, keys bool) *string {
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
			sb.WriteString("[")
			i:= 0
			for k, v := range pElem.GetKey() {
				sb.WriteString(k)
				sb.WriteString("=")
				sb.WriteString(v)
				if i != len(pElem.GetKey())-1 {
					sb.WriteString(",")
				}
			}
			sb.WriteString("]")
		}
		if i+1 != len(path.GetElem()) {
			sb.WriteString("/")
		}
	}
	return utils.StringPtr("/" + sb.String())
}

// XpathToGnmiPath convertss a xpath string to a config gnmi path
func (p *Parser) XpathToGnmiPath(xpath string, offset int) (path *gnmi.Path) {
	split := strings.Split(xpath, "/")
	for i, element := range split {
		// ignore the first element
		//fmt.Printf("i = %d, element = %s\n", i, element)
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

// TransformPathToLeafRefPath returns a config gnmi path tailored for leafrefs
// For a leafRef path the last entry of the name should be a key in the previous element
func (p *Parser) TransformGnmiPathToLeafRefPath(path *gnmi.Path) *gnmi.Path {
	key := path.GetElem()[len(path.GetElem())-1].Name
	path.Elem = path.Elem[:(len(path.GetElem()) - 1)]
	path.GetElem()[len(path.GetElem())-1].Key = make(map[string]string)
	path.GetElem()[len(path.GetElem())-1].Key[key] = ""
	return path
}

// TransformPathAsRelative2Resource returns a relative path
func (p *Parser) TransformGnmiPathAsRelative2Resource(localPath, activeResPath *gnmi.Path) *gnmi.Path {
	if len(activeResPath.GetElem()) >= 1 {
		localPath.Elem = localPath.Elem[(len(activeResPath.GetElem()) - 1):(len(localPath.GetElem()))]
	} else {
		localPath.Elem = localPath.Elem[:len(localPath.GetElem())]
	}

	return localPath
}

// AppendElemInPath adds a pathElem to the config gnmi path
func (p *Parser) AppendElemInGnmiPath(path *gnmi.Path, name string, keys []string) *gnmi.Path {
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

// AppendElemInPath adds a pathElem to the config gnmi path
func (p *Parser) AppendElemInGnmiPathWithFullKey(path *gnmi.Path, name string, key map[string]string) *gnmi.Path {
	pathElem := &gnmi.PathElem{
		Name: name,
	}
	if key != nil {
		pathElem.Key = key
	}

	path.Elem = append(path.Elem, pathElem)
	return path
}

func (p *Parser) CopyPathElemKey(key map[string]string) map[string]string {
	newKey := make(map[string]string)
	for k, v := range key {
		newKey[k] = v
	}
	return newKey
}

func (p *Parser) GetRemoteGnmiPathsFromResolvedLeafRef(resolvedLeafRef *ResolvedLeafRefGnmi) []*gnmi.Path {
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
					newKey := p.CopyPathElemKey(pathElem.GetKey())
					p.AppendElemInGnmiPathWithFullKey(remotePath, pathElem.GetName(), newKey)
					// we stop at copying the first key
					break
				} else {
					p.AppendElemInGnmiPathWithFullKey(remotePath, pathElem.GetName(), nil)
				}
			}
			if p.log != nil {
				p.log.Debug("GetRemotePathsFromResolvedLeafRef", "remotePath", remotePath)
			}
			remotePaths = append(remotePaths, remotePath)
		} else {
			remotePaths = append(remotePaths, resolvedLeafRef.RemotePath)
		}
	}
	return remotePaths
}

// RemoveFirstEntry removes the first entry of the xpath, so it trims the first element of the /
func (p *Parser) RemoveFirstEntry(s string) string {
	split := strings.Split(s, "/")
	var path string
	for i, s := range split {
		if i > 1 {
			path += "/" + s
		}
	}
	return path
}

// GetValueType return if a value is a slice or not
func (p *Parser) GetValueType(value interface{}) string {
	switch v := value.(type) {
	case map[string]interface{}:
		for _, v1 := range v {
			switch v1.(type) {
			case []interface{}:
				return Slice
			}
		}
	}
	return NonSlice
}

// GetKeyInfo returns all keys and values in a []slice
func (p *Parser) GetKeyInfo(keys map[string]string) ([]string, []string) {
	keyName := make([]string, 0)
	keyValue := make([]string, 0)
	for k, v := range keys {
		keyName = append(keyName, k)
		keyValue = append(keyValue, v)
	}
	return keyName, keyValue
}

// GetValue return the data of the gnmi typed value
func (p *Parser) GetValue(updValue *gnmi.TypedValue) (interface{}, error) {
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

func (p *Parser) DeepCopyGnmiPath(in *gnmi.Path) *gnmi.Path {
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

// CompareConfigPathsWithResourceKeys returns changed true when resourceKeys were provided
// and if they are different. In this case the deletePath is also valid, otherwise when changd is false
// the delete path is not reliable
func (p *Parser) CompareGnmiPathsWithResourceKeys(path *gnmi.Path, resourceKeys map[string]string) (bool, []*gnmi.Path, map[string]string) {
	changed := false
	deletePaths := make([]*gnmi.Path, 0)
	deletePath := &gnmi.Path{
		Elem: make([]*gnmi.PathElem, 0),
	}
	newKeys := make(map[string]string)
	for _, pathElem := range path.GetElem() {
		elem := &gnmi.PathElem{
			Name: pathElem.GetName(),
		}
		if len(pathElem.GetKey()) != 0 {
			elem.Key = make(map[string]string)
			for keyName, keyValue := range pathElem.GetKey() {
				if len(resourceKeys) != 0 {
					// the resource keys exists; if they dont exist there is no point comparing
					// the data
					if value, ok := resourceKeys[pathElem.GetName()+":"+keyName]; ok {
						if value != keyValue {
							changed = true
						}
						// use the value of the resourceKeys if the path should be deleted
						elem.Key[keyName] = value
					}
				}
				// these are the new keys which were supplied by the resource
				newKeys[pathElem.GetName()+":"+keyName] = keyValue
			}
		}
		deletePath.Elem = append(deletePath.Elem, elem)
	}
	deletePaths = append(deletePaths, deletePath)
	return changed, deletePaths, newKeys
}
