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
	"errors"
	"fmt"
	"strconv"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-yang/pkg/yentry"
)

// GetGranularUpdatesFromJSON provides an update per leaf level
func GetGranularUpdatesFromJSON(p *gnmi.Path, d interface{}, rs yentry.Handler) ([]*gnmi.Update, error) {
	u := make([]*gnmi.Update, 0)
	var err error
	u, err = getGranularUpdatesFromJSON(p, d, u, rs)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getGranularUpdatesFromJSON provides an update per leaf level
func getGranularUpdatesFromJSON(p *gnmi.Path, d interface{}, u []*gnmi.Update, rs yentry.Handler) ([]*gnmi.Update, error) {
	switch x := d.(type) {
	case map[string]interface{}:
		// add the keys as data in the last element
		for k, v := range p.GetElem()[len(p.GetElem())-1].GetKey() {
			value, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			u = append(u, &gnmi.Update{
				Path: &gnmi.Path{Elem: append(p.GetElem(), &gnmi.PathElem{Name: k})},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: value}},
			})
		}

		// add the values and add further processing
		for k, v := range x {
			switch val := v.(type) {
			case []interface{}:
				for _, v := range val {
					switch value := v.(type) {
					case map[string]interface{}:
						// gets the keys from the yangschema based on the gnmi path
						keys := rs.GetKeys(&gnmi.Path{
							Elem: append(p.GetElem(), &gnmi.PathElem{Name: k}),
						})
						// get the gnmi path with the key data
						newPath, err := getPathWithKeys(p, keys, k, value)
						if err != nil {
							return nil, err
						}
						u, err = getGranularUpdatesFromJSON(newPath, v, u, rs)
						if err != nil {
							return nil, err
						}
					}
				}
			default:
				// this would be map[string]interface{}
				// or string, other types
				value, err := json.Marshal(v)
				if err != nil {
					return nil, err
				}
				u = append(u, &gnmi.Update{
					Path: &gnmi.Path{Elem: append(p.GetElem(), &gnmi.PathElem{Name: k})},
					Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: value}},
				})
			}
		}

	}
	return u, nil
}

// GetUpdatesFromJSON provides an update per container, list and leaflist level
func GetUpdatesFromJSON(p *gnmi.Path, d interface{}, rs yentry.Handler) ([]*gnmi.Update, error) {
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
func getUpdatesFromJSON(p *gnmi.Path, d interface{}, u []*gnmi.Update, rs yentry.Handler) ([]*gnmi.Update, error) {
	var err error
	switch x := d.(type) {
	case map[string]interface{}:
		value := make(map[string]interface{})
		for k, v := range x {
			switch val := v.(type) {
			case []interface{}:
				for _, v := range val {
					switch value := v.(type) {
					case map[string]interface{}:
						// gets the keys from the yangschema based on the gnmi path
						keys := rs.GetKeys(&gnmi.Path{
							Elem: append(p.GetElem(), &gnmi.PathElem{Name: k}),
						})
						// get the gnmipath with the key data
						newPath, err := getPathWithKeys(p, keys, k, value)
						if err != nil {
							return nil, err
						}
						u, err = getUpdatesFromJSON(newPath, v, u, rs)
						if err != nil {
							return nil, err
						}
					}
				}
			case map[string]interface{}:
				// yang new container -> we provide a dedicated update
				u, err = getUpdatesFromJSON(
					&gnmi.Path{
						Elem: append(p.GetElem(), &gnmi.PathElem{Name: k}),
					}, val, u, rs)
				if err != nil {
					return nil, err
				}
			default:
				// string, other types -> we are at the end of the path
				// collect all the data for further processing
				value[k] = v
			}
		}
		// update for all the values in the container
		// adds the keys to the path and deletes them from the data/json
		if len(value) > 0 {
			update, err := getUpdatesFromContainer(p, value)
			if err != nil {
				return nil, err
			}
			u = append(u, update)
		}
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
func getUpdatesFromContainer(p *gnmi.Path, value map[string]interface{}) (*gnmi.Update, error) {
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
		Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: v}},
	}, nil
}