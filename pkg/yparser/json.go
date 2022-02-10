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
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-runtime/pkg/utils"
	"github.com/yndd/ndd-yang/pkg/yentry"
)

type updates struct {
	upds []*gnmi.Update
}

// GetGranularUpdatesFromJSON provides an update per leaf level
func GetGranularUpdatesFromJSON(p *gnmi.Path, d interface{}, rs *yentry.Entry) ([]*gnmi.Update, error) {
	var err error
	updates := &updates{
		upds: make([]*gnmi.Update, 0),
	}
	err = getGranularUpdatesFromJSON(p, d, updates, rs)
	if err != nil {
		return nil, err
	}
	/*
		for _, u := range updates.upds {
			fmt.Printf("GetGranularUpdatesFromJSON path: %s, value: %v\n", GnmiPath2XPath(u.Path, true), u.GetVal())
		}
	*/
	return updates.upds, nil
}

// getGranularUpdatesFromJSON provides an update per leaf level
func getGranularUpdatesFromJSON(path *gnmi.Path, d interface{}, u *updates, rs *yentry.Entry) error {
	//fmt.Printf("getGranularUpdatesFromJSON entry: path: %s, data: %v\n", GnmiPath2XPath(path, true), d)
	p := DeepCopyGnmiPath(path)

	pathKeys := make(map[string]string)
	if len(p.GetElem()) != 0 {
		// add the keys as data in the last element
		for k, v := range p.GetElem()[len(p.GetElem())-1].GetKey() {
			p := DeepCopyGnmiPath(p)

			value, err := json.Marshal(v)
			if err != nil {
				return err
			}
			u.upds = append(u.upds, &gnmi.Update{
				Path: &gnmi.Path{Elem: append(p.GetElem(), &gnmi.PathElem{Name: k})},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonIetfVal{JsonIetfVal: value}},
			})
			//fmt.Printf("getGranularUpdatesFromJSON key: path: %s, data:%v\n", GnmiPath2XPath(u.upds[len(u.upds)-1].Path, true), u.upds[len(u.upds)-1].Val)
		}

		pathKeys = p.GetElem()[len(p.GetElem())-1].GetKey()
		if len(pathKeys) == 0 {
			pathKeys = make(map[string]string)
		}
	}

	// process the data
	switch x := d.(type) {
	case map[string]interface{}:
		// add the values and add further processing
		for k, v := range x {
			if _, ok := pathKeys[k]; !ok {
				// the keys are already added before so we can ignore them
				// we process only data that is not added before
				p := DeepCopyGnmiPath(p)

				switch val := v.(type) {
				case []interface{}:
					leaflist := false
					for _, vval := range val {
						switch value := vval.(type) {
						case map[string]interface{}:
							// gets the keys from the yangschema based on the gnmi path
							//fmt.Printf("pathElem: %v\n", append(p.GetElem(), &gnmi.PathElem{Name: k}))
							keys := rs.GetKeys(&gnmi.Path{
								Elem: append(p.GetElem(), &gnmi.PathElem{Name: k}),
							})
							// get the gnmi path with the key data
							newPath, err := getPathWithKeys(DeepCopyGnmiPath(p), keys, k, value)
							if err != nil {
								return err
							}
							err = getGranularUpdatesFromJSON(newPath, vval, u, rs)
							if err != nil {
								return err
							}
						default: // leaf-list
							leaflist = true
						}
					}
					if leaflist {
						// leaflists are added as a single value
						value, err := json.Marshal(val)
						if err != nil {
							return err
						}
						u.upds = append(u.upds, &gnmi.Update{
							Path: &gnmi.Path{Elem: append(p.GetElem(), &gnmi.PathElem{Name: k})},
							Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonIetfVal{JsonIetfVal: value}},
						})
					}
				case map[string]interface{}:
					newPath := DeepCopyGnmiPath(p)
					newPath.Elem = append(newPath.GetElem(), &gnmi.PathElem{Name: k})
					err := getGranularUpdatesFromJSON(newPath, v, u, rs)
					if err != nil {
						return err
					}
				default:
					// this would be map[string]interface{}
					// or string, other types
					value, err := json.Marshal(v)
					if err != nil {
						return err
					}
					u.upds = append(u.upds, &gnmi.Update{
						Path: &gnmi.Path{Elem: append(p.GetElem(), &gnmi.PathElem{Name: k})},
						Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonIetfVal{JsonIetfVal: value}},
					})
					//fmt.Printf("getGranularUpdatesFromJSON default: path: %s, data:%v\n", GnmiPath2XPath(u.upds[len(u.upds)-1].Path, true), u.upds[len(u.upds)-1].Val)
				}
			}
		}
	}
	return nil
}

// GetUpdatesFromJSON provides an update per container, list and leaflist level
func GetUpdatesFromJSON(p *gnmi.Path, d interface{}, rs *yentry.Entry) ([]*gnmi.Update, error) {
	u := make([]*gnmi.Update, 0)
	var err error
	u, err = getUpdatesFromJSON(p, d, u, rs)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUpdatesFromJSON creates a gnmi update
// every list and container is put in a seperate update
func getUpdatesFromJSON(p *gnmi.Path, d interface{}, u []*gnmi.Update, rs *yentry.Entry) ([]*gnmi.Update, error) {
	var err error
	switch x := d.(type) {
	case map[string]interface{}:
		value := make(map[string]interface{})
		for k, v := range x {
			//fmt.Printf("getUpdatesFromJSON map[string]interface{}: path: %s, k:%s, v: %v\n", GnmiPath2XPath(p, true), k, v)
			switch val := v.(type) {
			case []interface{}:
				//fmt.Printf("getUpdatesFromJSON []interface{}: path: %s, k:%s, v: %v\n", GnmiPath2XPath(p, true), k, v)
				leaflist := false
				for _, v := range val {
					switch vv := v.(type) {
					case map[string]interface{}:
						// gets the keys from the yangschema based on the gnmi path
						keys := rs.GetKeys(&gnmi.Path{
							Elem: append(p.GetElem(), &gnmi.PathElem{Name: k}),
						})
						//fmt.Printf("getUpdatesFromJSON []interface{} keys: %v\n", keys)
						// get the gnmipath with the key data
						newPath, err := getPathWithKeys(DeepCopyGnmiPath(p), keys, k, vv)
						if err != nil {
							return nil, err
						}
						u, err = getUpdatesFromJSON(newPath, v, u, rs)
						if err != nil {
							return nil, err
						}
					default: // leaf-list
						leaflist = true
					}

				}
				if leaflist {
					// leaflists are added as a single value
					v, err := json.Marshal(val)
					if err != nil {
						return nil, err
					}
					u = append(u, &gnmi.Update{
						Path: &gnmi.Path{
							Elem: append(p.GetElem(), &gnmi.PathElem{Name: k}),
						},
						Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonIetfVal{JsonIetfVal: v}},
					})
				}
			case map[string]interface{}:
				// yang new container -> we provide a dedicated update
				//fmt.Printf("getUpdatesFromJSON map[string]interface{}: path: %s, k:%s, v: %v\n", GnmiPath2XPath(p, true), k, v)
				u, err = getUpdatesFromJSON(
					&gnmi.Path{
						Elem: append(p.GetElem(), &gnmi.PathElem{Name: k}),
					}, v, u, rs)
				if err != nil {
					return nil, err
				}
			default:
				//fmt.Printf("getUpdatesFromJSON default: path: %s, k:%s, v: %v\n", GnmiPath2XPath(p, true), k, v)
				// string, other types -> we are at the end of the path
				// collect all the data for further processing
				value[k] = v
			}
		}
		// update for all the values in the container
		// adds the keys to the path and deletes them from the data/json
		//if len(value) >= 0 {
		update, err := getUpdatesFromContainer(p, value)
		if err != nil {
			return nil, err
		}
		u = append(u, update)
		/*
			for _, upd := range u {
				fmt.Printf("getUpdatesFromJSON update, path: %s, val: %v\n", GnmiPath2XPath(upd.GetPath(), true), upd.GetVal())
			}
			//}
		*/
	}
	return u, nil
}

// getPathWithKeys provides a new path with the key data
func getPathWithKeys(p *gnmi.Path, keys []string, k string, value map[string]interface{}) (*gnmi.Path, error) {
	if len(keys) != 0 {
		pathKeys := make(map[string]string)
		for _, key := range keys {
			pathKeys[key] = fmt.Sprintf("%v", value[key])
		}
		return &gnmi.Path{
			Elem: append(p.GetElem(), &gnmi.PathElem{
				Name: k,
				Key:  pathKeys,
			})}, nil
	}
	// we should never come here
	return nil, errors.New("[]interface{} without keys is not expected")
	//newPath = &gnmi.Path{
	//	Elem: append(p.GetElem(), &gnmi.PathElem{Name: k}),
	//}
}

// getUpdatesFromContainer
// adds the keys to the path and deletes them from the data/json
func getUpdatesFromContainer(path *gnmi.Path, value map[string]interface{}) (*gnmi.Update, error) {
	p := DeepCopyGnmiPath(path)
	if len(p.GetElem()) > 0 {
		// if the path contains a key we need to remove the element from the value and add it in the path
		if len(p.GetElem()[len(p.GetElem())-1].GetKey()) != 0 {
			for k := range p.GetElem()[len(p.GetElem())-1].GetKey() {
				if v, ok := value[k]; ok {
					// add Value to path
					switch v := v.(type) {
					case string:
						p.GetElem()[len(p.GetElem())-1].GetKey()[k] = string(v)
					case uint32:
						p.GetElem()[len(p.GetElem())-1].GetKey()[k] = strconv.Itoa(int(v))
					case float64:
						p.GetElem()[len(p.GetElem())-1].GetKey()[k] = fmt.Sprintf("%.0f", v)
					}
					// delete element from the value
					delete(value, k)
				}
			}
		}
	}
	v, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return &gnmi.Update{
		Path: p,
		Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonIetfVal{JsonIetfVal: v}},
	}, nil
}

// GetHierIDsFromPath from path gets the hierarchical ids from the gnmi path
func GetHierIDsFromPath(p *gnmi.Path) []string {
	// get the hierarchical ids from the path
	hids := make([]string, 0)
	for i, pathElem := range p.GetElem() {
		if i < len(p.GetElem())-1 {
			if len(pathElem.GetKey()) != 0 {
				for k := range pathElem.GetKey() {
					hids = append(hids, pathElem.GetName()+"-"+k)
				}
			}
		}
	}
	return hids
}

// RemoveHierIDs removes the hierarchical IDs from the data
func RemoveHierIDsFomData(hids []string, x interface{}) interface{} {
	switch x := x.(type) {
	case map[string]interface{}:
		if len(hids) != 0 {
			for _, hid := range hids {
				delete(x, hid)
			}
		}
	}
	return x
}

func AddDataToList(x interface{}) (interface{}, error) {
	x1 := make(map[string]interface{})
	switch x := x.(type) {
	case map[string]interface{}:
		for k1, v1 := range x {
			x2 := make([]interface{}, 0)
			x2 = append(x2, v1)
			x1[k1] = x2
		}
		return x1, nil
	}

	// wrong data input
	return x1, errors.New(fmt.Sprintf("data transformation, wrong data input %v", x))

}

// CleanConfig2String returns a clean config and a string
// clean means removing the prefixes in the json elements
func CleanConfig2String(cfg map[string]interface{}) (map[string]interface{}, *string, error) {
	// trim the first map
	for _, v := range cfg {
		cfg = CleanConfig(v.(map[string]interface{}))
	}
	//fmt.Printf("cleanConfig Config %v\n", cfg)

	jsonConfigStr, err := json.Marshal(cfg)
	if err != nil {
		return nil, nil, err
	}
	return cfg, utils.StringPtr(string(jsonConfigStr)), nil
}

func CleanConfig(x1 map[string]interface{}) map[string]interface{} {
	x2 := make(map[string]interface{})
	for k1, v1 := range x1 {
		//fmt.Printf("cleanConfig Key: %s, Value: %v\n", k1, v1)
		switch x3 := v1.(type) {
		case []interface{}:
			x := make([]interface{}, 0)
			for _, v3 := range x3 {
				switch x3 := v3.(type) {
				case map[string]interface{}:
					x4 := CleanConfig(x3)
					x = append(x, x4)
				default:
					// clean the data
					switch v4 := v3.(type) {
					case string:
						x = append(x, strings.Split(v4, ":")[len(strings.Split(v4, ":"))-1])
					default:
						//fmt.Printf("type in []interface{}: %v\n", reflect.TypeOf(v4))
						x = append(x, v4)
					}
				}
			}
			x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = x
		case map[string]interface{}:
			x4 := CleanConfig(x3)
			x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = x4
		case string:
			// for string values there can be also a header in the values e.g. type, Value: srl_nokia-network-instance:ip-vrf
			if strings.Contains(x3, "::") {
				// avoids splitting ipv6 addresses
				x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = x3
			} else {
				// if there are more ":" in the string it is likely an esi or mac address
				if len(strings.Split(x3, ":")) <= 2 {
					x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = strings.Split(x3, ":")[len(strings.Split(x3, ":"))-1]
				} else {
					x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = x3
				}
			}
		case float64:
			x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = x3
		case bool:
			x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = x3

		default:
			// for other values like bool, float64, uint32 we dont do anything
			if x3 != nil {
				fmt.Printf("type in main: %v\n", reflect.TypeOf(x3))
			} else {
				fmt.Printf("type in main: nil\n")
			}
			x2[strings.Split(k1, ":")[len(strings.Split(k1, ":"))-1]] = x3
		}
	}
	return x2
}