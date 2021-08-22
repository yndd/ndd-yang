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

	config "github.com/netw-device-driver/ndd-grpc/config/configpb"
	"github.com/netw-device-driver/ndd-runtime/pkg/utils"
	"github.com/openconfig/gnmi/proto/gnmi"
)

// GnmiPath2ConfigPath converts the gnmi path to config path
func GnmiPath2ConfigPath(inPath *gnmi.Path) *config.Path {
	outPath := &config.Path{}
	outPath.Elem = make([]*config.PathElem, 0)
	if inPath != nil {
		for _, pElem := range inPath.GetElem() {
			elem := &config.PathElem{}
			elem.Name = strings.Split(pElem.GetName(), ":")[len(strings.Split(pElem.GetName(), ":"))-1]
			if len(pElem.GetKey()) != 0 {
				elem.Key = make(map[string]string)
				for key, value := range pElem.GetKey() {
					if strings.Contains(value, "::") {
						// avoids splitting ipv6 addresses
						elem.Key[strings.Split(key, ":")[len(strings.Split(key, ":"))-1]] = value
					} else {
						elem.Key[strings.Split(key, ":")[len(strings.Split(key, ":"))-1]] = strings.Split(value, ":")[len(strings.Split(value, ":"))-1]
					}

				}
			}
			outPath.Elem = append(outPath.Elem, elem)
		}
	}
	return outPath
}

// ConfigGnmiPathToName converts a config gnmi path to a name where each element of the
// path is seperated by a "-"
func ConfigGnmiPathToName(path *config.Path) string {
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
func GnmiPathToXPath(path *gnmi.Path, keys bool) *string {
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
			for k, v := range pElem.GetKey() {
				sb.WriteString("[")
				sb.WriteString(k)
				sb.WriteString("=")
				sb.WriteString(v)
				sb.WriteString("]")
			}
		}
		if i+1 != len(path.GetElem()) {
			sb.WriteString("/")
		}
	}
	return utils.StringPtr("/" + sb.String())
}

// ConfigGnmiPathToXPath converts a config gnmi path with or withour keys to a string pointer
func ConfigGnmiPathToXPath(path *config.Path, keys bool) *string {
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
			for k, v := range pElem.GetKey() {
				sb.WriteString("[")
				sb.WriteString(k)
				sb.WriteString("=")
				sb.WriteString(v)
				sb.WriteString("]")
			}
		}
		if i+1 != len(path.GetElem()) {
			sb.WriteString("/")
		}
	}
	return utils.StringPtr("/" + sb.String())
}

// XpathToGnmiPath convertss a xpath string to a config gnmi path
func XpathToGnmiPath(p string, offset int) (path *config.Path) {
	split := strings.Split(p, "/")
	for i, element := range split {
		// ignore the first element
		//fmt.Printf("i = %d, element = %s\n", i, element)
		if i == 0 {
			path = &config.Path{
				Elem: make([]*config.PathElem, 0),
			}
		} else {
			// offset is used to ignore an element from the path
			if i > offset {
				pathElem := &config.PathElem{}
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
func TransformPathToLeafRefPath(path *config.Path) *config.Path {
	key := path.GetElem()[len(path.GetElem())-1].Name
	path.Elem = path.Elem[:(len(path.GetElem()) - 1)]
	path.GetElem()[len(path.GetElem())-1].Key = make(map[string]string)
	path.GetElem()[len(path.GetElem())-1].Key[key] = ""
	return path
}

// TransformPathAsRelative2Resource returns a relative path
func TransformPathAsRelative2Resource(localPath, activeResPath *config.Path) *config.Path {
	localPath.Elem = localPath.Elem[(len(activeResPath.GetElem()) - 1):(len(localPath.GetElem()))]
	return localPath
}

// AppendElemInPath adds a pathElem to the config gnmi path
func AppendElemInPath(path *config.Path, name, key string) *config.Path {
	pathElem := &config.PathElem{
		Name: name,
	}
	if key != "" {
		pathElem.Key = make(map[string]string)
		pathElem.Key[key] = ""
	}

	path.Elem = append(path.Elem, pathElem)
	return path
}

// RemoveFirstEntry removes the first entry of the xpath, so it trims the first element of the /
func RemoveFirstEntry(s string) string {
	split := strings.Split(s, "/")
	var p string
	for i, s := range split {
		if i > 1 {
			p += "/" + s
		}
	}
	return p
}

// GetValueType return if a value is a slice or not
func GetValueType(value interface{}) string {
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
func GetKeyInfo(keys map[string]string) ([]string, []string) {
	keyName := make([]string, 0)
	keyValue := make([]string, 0)
	for k, v := range keys {
		keyName = append(keyName, k)
		keyValue = append(keyValue, v)
	}
	return keyName, keyValue
}

// GetValue return the data of the gnmo typed value
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
