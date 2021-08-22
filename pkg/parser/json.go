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
	"strings"

	"github.com/netw-device-driver/ndd-runtime/pkg/utils"
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
