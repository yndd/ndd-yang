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
	"fmt"
	"reflect"
	"strconv"
	"strings"

	config "github.com/netw-device-driver/ndd-grpc/config/configpb"
	"github.com/netw-device-driver/ndd-runtime/pkg/utils"
	"github.com/pkg/errors"
	"github.com/wI2L/jsondiff"
)

// Make a deep copy from in into out object.
func DeepCopy(in interface{}) (interface{}, error) {
	if in == nil {
		return nil, errors.New("in cannot be nil")
	}
	//fmt.Printf("json copy input %v\n", in)
	bytes, err := json.Marshal(in)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal input data")
	}
	var out interface{}
	err = json.Unmarshal(bytes, &out)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal to output data")
	}
	//fmt.Printf("json copy output %v\n", out)
	return out, nil
}

// RemoveHierarchicalKeys removes the hierarchical keys from the data
func RemoveHierarchicalKeys(d []byte, hids []string) ([]byte, error) {
	var x map[string]interface{}
	json.Unmarshal(d, &x)

	fmt.Printf("data before hierarchical key removal: %v\n", x)
	// we first go over hierarchical ids since when they are empty it optimizes the processing
	for _, h := range hids {
		for k := range x {
			if k == h {
				delete(x, k)
			}
		}
	}
	fmt.Printf("data after hierarchical key removal: %v\n", x)
	return json.Marshal(x)
}

// CleanConfig2String returns a clean config and a string
// clean means removing the prefixes in the json elements
func CleanConfig2String(cfg map[string]interface{}) (map[string]interface{}, *string, error) {
	// trim the first map
	for _, v := range cfg {
		cfg = cleanConfig(v.(map[string]interface{}))
	}
	fmt.Printf("cleanConfig Config %v\n", cfg)

	jsonConfigStr, err := json.Marshal(cfg)
	if err != nil {
		return nil, nil, err
	}
	return cfg, utils.StringPtr(string(jsonConfigStr)), nil
}

func cleanConfig(x1 map[string]interface{}) map[string]interface{} {
	x2 := make(map[string]interface{})
	for k1, v1 := range x1 {
		fmt.Printf("cleanConfig Key: %s, Value: %v\n", k1, v1)
		switch x3 := v1.(type) {
		case []interface{}:
			x := make([]interface{}, 0)
			for _, v3 := range x3 {
				switch x3 := v3.(type) {
				case map[string]interface{}:
					x4 := cleanConfig(x3)
					x = append(x, x4)
				default:
					// clean the data
					switch v4 := v3.(type) {
					case string:
						x = append(x, strings.Split(v4, ":")[len(strings.Split(v4, ":"))-1])
					default:
						fmt.Printf("type in []interface{}: %v\n", reflect.TypeOf(v4))
						x = append(x, v4)
					}
				}
			}
			x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = x
		case map[string]interface{}:
			x4 := cleanConfig(x3)
			x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = x4
		case string:
			// for string values there can be also a header in the values e.g. type, Value: srl_nokia-network-instance:ip-vrf
			x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = strings.Split(x3, ":")[len(strings.Split(x3, ":"))-1]
		default:
			// for other values like bool, float64, uint32 we dont do anything
			fmt.Printf("type in main: %v\n", reflect.TypeOf(x3))
			x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = v1
		}
	}
	return x2
}

// ParseTreeWithAction parses various actions on a json object in a recursive way
// actions can be Get, Update, Delete and Create
func ParseTreeWithAction(x1 interface{}, tc *TraceCtxt, idx int) interface{} {
	// idx is a local counter that will stay local, after the recurssive function calls it remains the same
	// tc.Idx is a global index used for tracing to trace, after a recursive function it will change if the recursive function changed it
	fmt.Printf("ParseTreeWithAction: %v, path: %v\n", tc, tc.Path)
	tc.Msg = append(tc.Msg, "entry")
	switch x1 := x1.(type) {
	case map[string]interface{}:
		tc.Msg = append(tc.Msg, "map[string]interface{}")
		if _, ok := x1[tc.Path.GetElem()[idx].GetName()]; ok {
			// object should exists
			tc.Msg = append(tc.Msg, "pathElem found")
			if idx == len(tc.Path.GetElem())-1 {
				if len(tc.Path.GetElem()[idx].GetKey()) != 0 {
					tc.Msg = append(tc.Msg, "end of path with key")
					// not last element of the list e.g. we are at interface of interface[name=ethernet-1/1]
					switch tc.Action {
					case ConfigTreeActionGet:
						return ParseTreeWithAction(x1[tc.Path.GetElem()[idx].GetName()], tc, idx)
					case ConfigTreeActionDelete:
						x1[tc.Path.GetElem()[idx].GetName()] = ParseTreeWithAction(x1[tc.Path.GetElem()[idx].GetName()], tc, idx)
						// if this is the last element in the slice we can delete the key from the list
						// e.g. delete subinterface[index=0] from interface[name=x] and it was the last subinterface in the interface
						switch x2 := x1[tc.Path.GetElem()[idx].GetName()].(type) {
						case []interface{}:
							if len(x2) == 0 {
								tc.Msg = append(tc.Msg, "removed last entry in the list with keys")
								delete(x1, tc.Path.GetElem()[idx].GetName())
							}
						}
						return x1
					case ConfigTreeActionCreate, ConfigTreeActionUpdate:
						x1[tc.Path.GetElem()[idx].GetName()] = ParseTreeWithAction(x1[tc.Path.GetElem()[idx].GetName()], tc, idx)
						return x1
					}
				} else {
					// system/ntp
					tc.Msg = append(tc.Msg, "end of path without key")
					tc.Found = true
					switch tc.Action {
					case ConfigTreeActionGet:
						return x1[tc.Path.GetElem()[idx].GetName()]
					case ConfigTreeActionDelete:
						delete(x1, tc.Path.GetElem()[idx].GetName())
						return x1
					case ConfigTreeActionUpdate:
						switch vv := tc.Value.(type) {
						case map[string]interface{}:
							for k, v := range vv {
								switch vvv := v.(type) {
								case string:
									x1[strings.Split(k, ":")[len(strings.Split(k, ":"))-1]] = strings.Split(vvv, ":")[len(strings.Split(vvv, ":"))-1]
								default:
									x1[strings.Split(k, ":")[len(strings.Split(k, ":"))-1]] = v
								}
							}
						}
						return x1
					case ConfigTreeActionCreate:
						x1[tc.Path.GetElem()[idx].GetName()] = copyAndCleanTxValues(tc.Value)
						return x1
					}

				}
			} else {
				if len(tc.Path.GetElem()[idx].GetKey()) != 0 {
					tc.Msg = append(tc.Msg, "not end of path with key")
					// not last element of the list e.g. we are at interface of interface[name=ethernet-1/1]/subinterface[index=100]
					switch tc.Action {
					case ConfigTreeActionGet:
						return ParseTreeWithAction(x1[tc.Path.GetElem()[idx].GetName()], tc, idx)
					case ConfigTreeActionDelete:
						x1[tc.Path.GetElem()[idx].GetName()] = ParseTreeWithAction(x1[tc.Path.GetElem()[idx].GetName()], tc, idx)
						return x1
					case ConfigTreeActionCreate, ConfigTreeActionUpdate:
						x1[tc.Path.GetElem()[idx].GetName()] = ParseTreeWithAction(x1[tc.Path.GetElem()[idx].GetName()], tc, idx)
						return x1
					}
				} else {
					// not last element of network-instance[name=ethernet-1/1]/protocol/bgp-vpn; we are at protocol level
					tc.Idx++
					tc.Msg = append(tc.Msg, "end of path without key")
					switch tc.Action {
					case ConfigTreeActionGet:
						return ParseTreeWithAction(x1[tc.Path.GetElem()[idx].GetName()], tc, idx+1)
					case ConfigTreeActionDelete, ConfigTreeActionCreate, ConfigTreeActionUpdate:
						x1[tc.Path.GetElem()[idx].GetName()] = ParseTreeWithAction(x1[tc.Path.GetElem()[idx].GetName()], tc, idx+1)
						return x1
					}
				}
			}
		}
		tc.Msg = append(tc.Msg, "map[string]interface{} not found")
		// this branch is mainly used for object creation
		switch tc.Action {
		case ConfigTreeActionDelete, ConfigTreeActionUpdate:
			// when the data is not found we just return x1 since nothing can get deleted or updated
			tc.Found = false
			tc.Data = x1
			return x1
		case ConfigTreeActionGet:
			tc.Found = false
			tc.Data = x1
			return x1
		case ConfigTreeActionCreate:
			// this branch is used to insert leafs, leaflists in the tree when object get created
			tc.Found = false
			if idx == len(tc.Path.GetElem())-1 {
				tc.Msg = append(tc.Msg, "map[string]interface{} last element in path, added item in the list")
				if len(tc.Path.GetElem()[idx].GetKey()) != 0 {
					tc.Msg = append(tc.Msg, "with key")
					// this is a new leaflist so we need to create the []interface
					// and add the key to map[string]interface{}
					// e.g. add subinterface[index=0] with value: admin-state: enable
					x2 := make([]interface{}, 0)
					// copy the values
					x3 := copyAndCleanTxValues(tc.Value)
					// add the keys to the list
					switch x4 := x3.(type) {
					case map[string]interface{}:
						// add the key of the path to the list
						for k, v := range tc.Path.GetElem()[idx].GetKey() {
							// add clean element to the list
							x4[strings.Split(k, ":")[len(strings.Split(k, ":"))-1]] = strings.Split(v, ":")[len(strings.Split(v, ":"))-1]
						}
						x2 = append(x2, x4)
					}
					x1[tc.Path.GetElem()[idx].GetName()] = x2
				} else {
					// create an mtu in
					tc.Msg = append(tc.Msg, "without key")
					x1[tc.Path.GetElem()[idx].GetName()] = copyAndCleanTxValues(tc.Value)
				}
				return x1
			} else {
				// it can be that we get a new creation with a path that is not fully created
				// e.g. /interface[name=ethernet-1/49]/subinterface[index=0]/vlan/encap/untagged
				//  and we only had /interface[name=ethernet-1/49]/subinterface[index=0] in the config
				tc.Msg = append(tc.Msg, "map[string]interface{} not last last element in path, adding element to the tree")
				tc.Idx++
				// create a new map string interface which will be recursively filled
				x1[tc.Path.GetElem()[idx].GetName()] = make(map[string]interface{})
				x1[tc.Path.GetElem()[idx].GetName()] = ParseTreeWithAction(x1[tc.Path.GetElem()[idx].GetName()], tc, idx+1)
				return x1
			}
		}
	case []interface{}:
		//fmt.Printf("ParseTreeWithAction []interface{}, idx: %d, path length %d, path: %v\n data: %v\n", idx, len(path.GetElem()), path.GetElem(), x1)
		tc.Msg = append(tc.Msg, "[]interface{}")
		for n, v := range x1 {
			switch x2 := v.(type) {
			case map[string]interface{}:
				if len(tc.Path.GetElem()[idx].GetKey()) != 0 {
					pathElemKeyNames, pathElemKeyValues := GetKeyInfo(tc.Path.GetElem()[idx].GetKey())
					tc.Msg = append(tc.Msg, fmt.Sprintf("pathElemKeyNames %v, pathElemKeyValues%v", pathElemKeyNames, pathElemKeyValues))
					// loop over all pathElemKeyNames
					// TODO multiple keys and values need to be updated !
					for i, pathElemKeyName := range pathElemKeyNames {
						if x3, ok := x2[pathElemKeyName]; ok {
							// pathElemKeyName found
							tc.Msg = append(tc.Msg, fmt.Sprintf("pathElemKeyName found: %s", pathElemKeyName))
							if idx == len(tc.Path.GetElem())-1 {
								tc.Msg = append(tc.Msg, "end of path with key")
								// last element in the pathElem list
								// example: interface[lag1] or interface[ethernet-1/1] is treated here
								switch x := x3.(type) {
								case string:
									//fmt.Printf("findObjectInTree string a: %s, b: %s\n", string(x), value)
									if string(x) == pathElemKeyValues[i] {
										//fmt.Printf("new data: x1 %v", x1)
										tc.Found = true
										tc.Msg = append(tc.Msg, fmt.Sprintf("pathElemKeyValue found: %s string", pathElemKeyValues[i]))
										switch tc.Action {
										case ConfigTreeActionGet:
											return x1[n]
										case ConfigTreeActionDelete:
											x1 = append(x1[:n], x1[n+1:]...)
											return x1
										case ConfigTreeActionUpdate:
											x1[n] = copyAndCleanTxValues(tc.Value)
											return x1
										case ConfigTreeActionCreate:
											// TODO if we ever come here
											return x1
										}
									}
								case uint32:
									//fmt.Printf("findObjectInTree uint32 a: %s, b: %s\n", strconv.Itoa(int(x)), value)
									if strconv.Itoa(int(x)) == pathElemKeyValues[i] {
										//fmt.Printf("new data: x1 %v", x1)
										tc.Found = true
										tc.Msg = append(tc.Msg, fmt.Sprintf("pathElemKeyValue found: %s uint32", pathElemKeyValues[i]))
										switch tc.Action {
										case ConfigTreeActionGet:
											return x1[n]
										case ConfigTreeActionDelete:
											x1 = append(x1[:n], x1[n+1:]...)
											return x1
										case ConfigTreeActionUpdate:
											x1[n] = copyAndCleanTxValues(tc.Value)
											return x1
										case ConfigTreeActionCreate:
											// TODO if we ever come here
											return x1
										}
									}
								case float64:
									//fmt.Printf("findObjectInTree float64 a: %s, b: %s\n", fmt.Sprintf("%.0f", x), value)
									if fmt.Sprintf("%.0f", x) == pathElemKeyValues[i] {
										//fmt.Printf("new data: x1 %v", x1)
										tc.Found = true
										tc.Msg = append(tc.Msg, fmt.Sprintf("pathElemKeyValue found: %s float64", pathElemKeyValues[i]))
										switch tc.Action {
										case ConfigTreeActionGet:
											return x1[n]
										case ConfigTreeActionDelete:
											x1 = append(x1[:n], x1[n+1:]...)
											return x1
										case ConfigTreeActionUpdate:
											x1[n] = copyAndCleanTxValues(tc.Value)
											return x1
										case ConfigTreeActionCreate:
											// TODO if we ever come here
											return x1
										}
									}
								default:
									tc.Found = false
									tc.Msg = append(tc.Msg, "[]interface{} pathElemKeyValue not found"+"."+fmt.Sprintf("%v", (reflect.TypeOf(x))))
									tc.Data = x1
									// we should not return here since there can be multiple entries in the list
									// e.g. interface[name=mgmt] and interface[name=etehrente-1/1]
									// we need to loop over all of them and the global for loop will return if not found
									//return x1
								}
							} else {
								// we hit this e.g. at interface level of interface[system0]/subinterface[index=0]
								tc.Msg = append(tc.Msg, "not end of path with key")
								switch x := x3.(type) {
								case string:
									if string(x) == pathElemKeyValues[i] {
										tc.Idx++
										tc.Msg = append(tc.Msg, fmt.Sprintf("pathElemKeyValue found: %s string", pathElemKeyValues[i]))
										switch tc.Action {
										case ConfigTreeActionGet:
											return ParseTreeWithAction(x1[n], tc, idx+1)
										case ConfigTreeActionDelete, ConfigTreeActionUpdate, ConfigTreeActionCreate:
											x1[n] = ParseTreeWithAction(x1[n], tc, idx+1)
											return x1
										}
									}
								case uint32:
									if strconv.Itoa(int(x)) == pathElemKeyValues[i] {
										tc.Idx++
										tc.Msg = append(tc.Msg, fmt.Sprintf("pathElemKeyValue found: %s uint32", pathElemKeyValues[i]))
										switch tc.Action {
										case ConfigTreeActionGet:
											return ParseTreeWithAction(x1[n], tc, idx+1)
										case ConfigTreeActionDelete, ConfigTreeActionUpdate, ConfigTreeActionCreate:
											x1[n] = ParseTreeWithAction(x1[n], tc, idx+1)
											return x1
										}
									}
								case float64:
									if fmt.Sprintf("%.0f", x) == pathElemKeyValues[i] {
										tc.Idx++
										tc.Msg = append(tc.Msg, fmt.Sprintf("pathElemKeyValue found: %s float64", pathElemKeyValues[i]))
										switch tc.Action {
										case ConfigTreeActionGet:
											return ParseTreeWithAction(x1[n], tc, idx+1)
										case ConfigTreeActionDelete, ConfigTreeActionUpdate, ConfigTreeActionCreate:
											x1[n] = ParseTreeWithAction(x1[n], tc, idx+1)
											return x1
										}
									}
								default:
									tc.Found = false
									tc.Msg = append(tc.Msg, "[]interface{} not found"+"."+fmt.Sprintf("%v %v", (reflect.TypeOf(x)), x))
									tc.Data = x1
									// we should not return here since there can be multiple entries in the list
									// e.g. interface[name=mgmt] and interface[name=etehrente-1/1]
									// we need to loop over all of them and the global for loop will return if not found
									//return x1
								}
							}
						} else {
							tc.Found = false
							tc.Data = x1
							tc.Msg = append(tc.Msg, fmt.Sprintf("pathElemKeyName not found: %s", pathElemKeyName))
						}
					}
				}
			}
		}
		tc.Msg = append(tc.Msg, "[]interface{} not found")
		// this is used to add an element to a list that already exists
		// e.g. interface[name=ethernet-1/49]/subinterface[index=0] exists and we add interface[name=ethernet-1/49]/subinterface[index=1]
		switch tc.Action {
		case ConfigTreeActionDelete, ConfigTreeActionUpdate, ConfigTreeActionGet:
			// when the data is not found we just return x1 since nothing can get deleted or updated
			tc.Found = false
			tc.Data = x1
			return x1
		case ConfigTreeActionCreate:
			if idx == len(tc.Path.GetElem())-1 {
				tc.Found = false
				tc.Data = x1
				tc.Msg = append(tc.Msg, "add element in an existing list")
				// copy the data of the information
				// add the key of the path to the data
				x3 := copyAndCleanTxValues(tc.Value)
				// add the keys to the list
				switch x4 := x3.(type) {
				case map[string]interface{}:
					// add the key of the path to the list
					for k, v := range tc.Path.GetElem()[idx].GetKey() {
						// add clean element to the list
						x4[strings.Split(k, ":")[len(strings.Split(k, ":"))-1]] = strings.Split(v, ":")[len(strings.Split(v, ":"))-1]
					}
					x1 = append(x1, x4)
				}
				return x1
			}
		}
	case nil:
	default:
	}
	switch tc.Action {
	case ConfigTreeActionDelete, ConfigTreeActionUpdate, ConfigTreeActionGet, ConfigTreeActionCreate:
		// when the data is not found we just return x1 since nothing can get deleted or updated
		tc.Found = false
		tc.Data = x1
		tc.Msg = append(tc.Msg, "default")
		return x1
	default:
		// when the data is not found we just return x1 since nothing can get deleted or updated
		tc.Found = false
		tc.Data = x1
		tc.Msg = append(tc.Msg, "default")
		return x1
	}
}

func copyAndCleanTxValues(value interface{}) interface{} {
	switch vv := value.(type) {
	case map[string]interface{}:
		x := make(map[string]interface{})
		for k, v := range vv {
			switch vvv := v.(type) {
			case string:
				x[strings.Split(k, ":")[len(strings.Split(k, ":"))-1]] = strings.Split(vvv, ":")[len(strings.Split(vvv, ":"))-1]
			default:
				x[strings.Split(k, ":")[len(strings.Split(k, ":"))-1]] = v
			}
		}
		return x
	case string:
		return strings.Split(vv, ":")[len(strings.Split(vv, ":"))-1]
	}
	return value
}

// CompareValues compares the 2 values and provides a json diff result
func CompareValues(path *config.Path, cacheValue, deviceValue interface{}, valueType string) (jsondiff.Patch, error) {
	x1, err := CleanCacheValueForComparison(path, cacheValue, valueType)
	if err != nil {
		return nil, err
	}
	x2, err := CleanDeviceValueForComparison(deviceValue)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("Data Comparison:\nx1: %v\nx2: %v\n", x1, x2)
	patch, err := jsondiff.Compare(x1, x2)
	if err != nil {
		return nil, errors.Wrap(err, ErrJSONCompare)
	}
	if len(patch) != 0 {
		fmt.Printf("Data Comparison failed:\nx1: %v\nx2: %v\n", x1, x2)
	}
	return patch, nil
}

// CleanDeviceValueForComparison cleans the data coming from the device
// it cleans the prefixes of the yang value; key and value
func CleanDeviceValueForComparison(deviceValue interface{}) (interface{}, error) {
	var x1 interface{}
	switch x := deviceValue.(type) {
	case map[string]interface{}:
		for k, v := range x {
			// if a string contains a : we return the last string after the :
			sk := strings.Split(k, ":")[len(strings.Split(k, ":"))-1]
			if k != sk {
				switch v.(type) {
				case string:
					v = strings.Split(fmt.Sprintf("%v", v), ":")[len(strings.Split(fmt.Sprintf("%v", v), ":"))-1]
				}
				delete(x, k)
				x[sk] = v
			} else {
				switch v.(type) {
				case string:
					v = strings.Split(fmt.Sprintf("%v", v), ":")[len(strings.Split(fmt.Sprintf("%v", v), ":"))-1]
				}
				x[sk] = v
			}
		}
		x1 = x
	}
	return x1, nil
}

// we update the cache value for comparison
// 1. any map[string]interface{} -> will come from another subscription
// 2. any key in the path can be removed since this is part of the path iso data comparison
// 3. if the value is a slice we should remove all strings/int/floats, if the data is not a slice we remove all slices
// -> the gnmi server splits slice data and non slice data
func CleanCacheValueForComparison(path *config.Path, cacheValue interface{}, valueType string) (x1 interface{}, err error) {
	// delete all leaftlists and keys of the cache data for comparison
	keyNames := make([]string, 0)
	if path.GetElem()[len(path.GetElem())-1].GetKey() != nil {
		keyNames, _ = GetKeyInfo(path.GetElem()[len(path.GetElem())-1].GetKey())
	}
	if cacheValue != nil {
		x1, err = DeepCopy(cacheValue)
		if err != nil {
			return nil, err
		}
	}
	switch x := x1.(type) {
	case map[string]interface{}:
		for k, v := range x {
			switch v.(type) {
			// delete maps since they come with a different xpath if present
			case map[string]interface{}:
				delete(x, k)
			// delete lists since they come with a different xpath if present
			case []interface{}:
				//fmt.Printf("cleanCacheValueForComparison valueType: %s", valueType)
				// if valuetype is a slice we should keep the slices, but delete the non slice information
				if valueType != Slice {
					delete(x, k)
				}
			case string:
				//fmt.Printf("cleanCacheValueForComparison valueType: %s", valueType)
				// loop over multiple keys
				if valueType != Slice {
					for _, keyName := range keyNames {
						if k == keyName {
							delete(x, k)
						}
					}
				} else {
					// when valuetype is a slice we should delete all regular entries
					delete(x, k)
				}
			case float64:
				// loop over multiple keys
				if valueType != Slice {
					for _, keyName := range keyNames {
						if k == keyName {
							delete(x, k)
						}
					}
				} else {
					// when valuetype is a slice we should delete all regular entries
					delete(x, k)
				}
			case bool:
				// loop over multiple keys
				if valueType != Slice {
					for _, keyName := range keyNames {
						if k == keyName {
							delete(x, k)
						}
					}
				} else {
					// when valuetype is a slice we should delete all regular entries
					delete(x, k)
				}
			case nil:
			default:
				// TODO add better logging
				fmt.Printf("cleanCacheValueForComparison Unknown type: %v\n", reflect.TypeOf(v))
			}
		}
		x1 = x
	}
	return x1, nil
}
