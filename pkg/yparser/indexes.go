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

import "github.com/openconfig/gnmi/proto/gnmi"

// CompareConfigPathsWithResourceKeys returns changed true when resourceKeys were provided
// and if they are different. In this case the deletePath is also valid, otherwise when changd is false
// the delete path is not reliable
func CompareGnmiPathsWithResourceKeys(rootPath *gnmi.Path, origResourceKeys map[string]string) (bool, []*gnmi.Path, map[string]string) {
	changed := false
	deletePaths := make([]*gnmi.Path, 0)
	deletePath := &gnmi.Path{
		Elem: make([]*gnmi.PathElem, 0),
	}
	newKeys := make(map[string]string)
	for _, pathElem := range rootPath.GetElem() {
		elem := &gnmi.PathElem{
			Name: pathElem.GetName(),
		}
		if len(pathElem.GetKey()) != 0 {
			elem.Key = make(map[string]string)
			for keyName, keyValue := range pathElem.GetKey() {
				if len(origResourceKeys) != 0 {
					// the resource keys exists; if they dont exist there is no point comparing
					// the data
					if value, ok := origResourceKeys[pathElem.GetName()+":"+keyName]; ok {
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
