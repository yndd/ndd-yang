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
	"fmt"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-yang/pkg/leafref"
	"github.com/yndd/ndd-yang/pkg/yentry"
)

// ValidateParentDependency validates the parent resource dependency
// based on the result this function returns the result + information on the validation
// we use a get here since we resolved the values of the keys alreay
func ValidateParentDependency(x1 interface{}, parentDependencies []*leafref.LeafRef, rs *yentry.Entry) (bool, []*leafref.ResolvedLeafRef, error) {
	// a global indication if the leafRef resolution was successfull or not
	// we are positive so we initialize to true
	success := true
	// we initialize a global list for finer information on the resolution
	resultValidations := make([]*leafref.ResolvedLeafRef, 0)
	// for all defined parent dependencies check if the remote leafref exists
	for _, leafRef := range parentDependencies {
		if len(leafRef.RemotePath.GetElem()) > 0 {
			fmt.Printf("ValidateParentDependency: %s\n", GnmiPath2XPath(leafRef.RemotePath, true))
			found := rs.IsPathPresent(&gnmi.Path{Elem: []*gnmi.PathElem{{}}}, leafRef.RemotePath, "", x1)
			if !found {
				success = false
			}
			resultValidation := &leafref.ResolvedLeafRef{
				LeafRef: &leafref.LeafRef{
					RemotePath: leafRef.RemotePath,
				},
				Value:    "",
				Resolved: found,
			}
			resultValidations = append(resultValidations, resultValidation)
		}
	}
	return success, resultValidations, nil
}
