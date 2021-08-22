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

	"github.com/pkg/errors"
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
