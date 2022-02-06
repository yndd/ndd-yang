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
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/pkg/errors"
)

const (
	errJSONMarshal       = "cannot marshal JSON object"
	errJSONUnMarshal     = "cannot unmarshal JSON object"
	errJSONCompare       = "cannot compare JSON objects"
	errJSONMarshalIndent = "cannot marshal JSON object with indent"
)

// A OperationType represents an operatio on a JSON resource
type OperationType string

// Condition Kinds.
const (
	// delete
	OperationTypeDelete OperationType = "Delete"
	// replace
	OperationTypeUpdate OperationType = "Update"
	// create
	OperationTypeCreate OperationType = "Create"
)

type Operation struct {
	Type  OperationType
	Path  string
	Value interface{}
}

func FindResourceDelta(updatesx1, updatesx2 []*gnmi.Update) ([]*gnmi.Path, []*gnmi.Update, error) {

	deletes := make([]*gnmi.Path, 0)
	updates := make([]*gnmi.Update, 0)
	// First we check if there are paths, which are created but should not be there!
	// We run over the data from the response and check if there are paths which are not found
	// in the intended data, if so we should delete the path
	for _, updatex2 := range updatesx2 {
		found := false
		for _, updatex1 := range updatesx1 {
			if GnmiPath2XPath(updatex1.Path, true) == GnmiPath2XPath(updatex2.Path, true) {
				found = true
			}
		}
		// path not found we should create it
		if !found {
			fmt.Printf("path not found in the intended data data x1: %s\n", GnmiPath2XPath(updatex2.Path, true))
			deletes = append(deletes, updatex2.Path)
		}
	}
	// After we check of the intended data is modified or deleted
	// We compare the intended data with the response data
	for _, updatex1 := range updatesx1 {
		found := false
		for _, updatex2 := range updatesx2 {
			if GnmiPath2XPath(updatex1.Path, true) == GnmiPath2XPath(updatex2.Path, true) {
				found = true
				//fmt.Printf("path x1: %s\n", *p.ConfigGnmiPathToXPath(updatex1.Path, true))
				//fmt.Printf("path x2: %s\n", *p.ConfigGnmiPathToXPath(updatex2.Path, true))
				//fmt.Printf("Spec Data: %v\n", updatex1.GetVal())
				//fmt.Printf("Resp Data: %v\n", updatex2.GetVal())
				x1, err := GetValue(updatex1.Val)
				if err != nil {
					return nil, nil, err
				}
				b1, err := json.Marshal(x1)
				if err != nil {
					return nil, nil, err
				}

				x2, err := GetValue(updatex2.Val)
				if err != nil {
					return nil, nil, err
				}
				b2, err := json.Marshal(x2)
				if err != nil {
					return nil, nil, err
				}
				patch, err := CompareJSONData(b1, b2)
				if err != nil {
					return nil, nil, errors.Wrap(err, errJSONMarshalIndent)
				}
				if len(patch) != 0 {
					// resource is not up to date
					fmt.Printf("Patch: %v\n", patch)
					for _, operation := range patch {

						v, err := json.Marshal(operation.Value)
						if err != nil {
							return nil, nil, err
						}
						fmt.Printf("Patch Operation: type %v, path: %v, value: %v\n", operation.Type, operation.Path, string(v))
						switch operation.Type {
						case OperationTypeDelete:
							path := &gnmi.Path{
								Elem: append(updatex1.GetPath().GetElem(), &gnmi.PathElem{Name: operation.Path}),
							}
							deletes = append(deletes, path)
						case OperationTypeUpdate:

							path := &gnmi.Path{
								Elem: append(updatex1.GetPath().GetElem(), &gnmi.PathElem{Name: operation.Path}),
							}

							//fmt.Printf("Patch Replace Data: %v\n", operation.Value)
							value, err := json.Marshal(operation.Value)
							if err != nil {
								return nil, nil, err
							}
							updates = append(updates, &gnmi.Update{
								Path: path,
								//Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonIetfVal{JsonIetfVal: value}},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_JsonIetfVal{
										JsonIetfVal: bytes.Trim(value, " \r\n\t"),
									},
								},
							})
						case OperationTypeCreate:
							// reapply the same data to the cache since we have individual paths
							// which means we can reapply the data
							path := DeepCopyGnmiPath(updatex1.Path)

							updates = append(updates, &gnmi.Update{
								Path: path,
								Val:  updatex1.Val,
							})

						default:
							fmt.Printf("Json Patch difference, Patch Operation: type %v, path: %v, value: %v\n", operation.Type, operation.Path, operation.Value)
						}
					}
					return deletes, updates, nil
				}
				continue
			}
		}
		// path not found we should create it
		if !found {
			fmt.Printf("path not found  in data x1: %s\n", GnmiPath2XPath(updatex1.Path, true))
			updates = append(updates, updatex1)
		}
	}

	if len(deletes) == 0 {
		//fmt.Printf("FindResourceDelta2 Delete up to date\n")
	} else {
		if len(deletes) == 2 {
			first := deletes[0]
			last := deletes[1]
			deletes[0] = last
			deletes[1] = first
		}
	}
	return deletes, updates, nil
}

// CompareJSONData compares the target with the source and provides operation guides
func CompareJSONData(t, s []byte) ([]Operation, error) {
	var x1, x2 interface{}
	if err := json.Unmarshal(t, &x1); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(s, &x2); err != nil {
		return nil, err
	}

	operations := make([]Operation, 0)
	switch xx1 := x1.(type) {
	case map[string]interface{}:
		switch xx2 := x2.(type) {
		case map[string]interface{}:
			// check for deletes, loop over the source object
			// see if all the elements exist in the target object
			for k2 := range xx2 {
				if _, ok := xx1[k2]; !ok {
					// key not found in the resource, delete it
					operations = append(operations, Operation{Type: OperationTypeDelete, Path: k2})
				}
			}
			// loop over the objects and check if
			// - the  elements exists or not -> if not add them
			// - if the data differs -> we add them
			for k1, v1 := range xx1 {
				if v2, ok := xx2[k1]; !ok {
					// the element does not exist
					fmt.Printf("Path OperationTypeUpdate element does not exist: xx2: %v, k1: %v, v1: %v\n", xx2, k1, v1)
					operations = append(operations, Operation{Type: OperationTypeUpdate, Path: k1, Value: v1})
				} else {
					// check if the value differs
					if v1 != v2 {
						// the data differs
						fmt.Printf("Path OperationTypeUpdate value differs v1: %v, v2: %v, type v1: %v, type v2: %v\n", v1, v2, reflect.TypeOf(v1), reflect.TypeOf(v2))
						operations = append(operations, Operation{Type: OperationTypeUpdate, Path: k1, Value: v1})
					}
				}
			}
		default:
			// we cannot compare the object so we should replace it
			fmt.Printf("CompareJSONDataCannot compare -> replace the object, type x1: %v, type x2: %v", reflect.TypeOf(x1), reflect.TypeOf(x2))
			for k1, v1 := range xx1 {
				operations = append(operations, Operation{Type: OperationTypeDelete, Path: k1, Value: v1})
			}
		}
	case []interface{}: //leaf-lists
		// we expect only 1 element presnt
		switch xx2 := x2.(type) {
		case []interface{}:
			for _, v1 := range xx1 {
				found := false
				for _, v2 := range xx2 {
					if v1 == v2 {
						found = true
						break
					}
				}
				if !found {
					operations = append(operations, Operation{Type: OperationTypeCreate})
				}
			}
		default:
			// data is not present
			operations = append(operations, Operation{Type: OperationTypeCreate})
		}
	default:
		// this is not an object but a string or float or integer instead
		fmt.Printf("CompareJSONData Default, type x1: %v, type x2: %v", reflect.TypeOf(x1), reflect.TypeOf(x2))
		// check if the value differs
		if x1 != x2 {
			// the data differs
			operations = append(operations, Operation{Type: OperationTypeCreate})
		}
	}
	return operations, nil
}
