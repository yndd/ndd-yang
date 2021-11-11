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

package leafref

import "github.com/openconfig/gnmi/proto/gnmi"



type ResolvedLeafRefGnmi struct {
 	LeafRef
	Value      string     `json:"value,omitempty"`
	Resolved   bool       `json:"resolved,omitempty"`
}

func (in *ResolvedLeafRefGnmi) DeepCopy() (out *ResolvedLeafRefGnmi) {
	out = new(ResolvedLeafRefGnmi)
	if in.LocalPath != nil {
		out.LocalPath = new(gnmi.Path)
		out.LocalPath.Elem = make([]*gnmi.PathElem, 0)
		for _, v := range in.LocalPath.GetElem() {
			elem := &gnmi.PathElem{}
			elem.Name = v.Name
			if len(v.GetKey()) != 0 {
				elem.Key = make(map[string]string)
				for key, value := range v.Key {
					elem.Key[key] = value
				}
			}
			out.LocalPath.Elem = append(out.LocalPath.Elem, elem)
		}
	}
	if in.RemotePath != nil {
		out.RemotePath = new(gnmi.Path)
		out.RemotePath.Elem = make([]*gnmi.PathElem, 0)
		for _, v := range in.RemotePath.GetElem() {
			elem := &gnmi.PathElem{}
			elem.Name = v.Name
			if len(v.GetKey()) != 0 {
				elem.Key = make(map[string]string)
				for key, value := range v.Key {
					elem.Key[key] = value
				}
			}
			out.RemotePath.Elem = append(out.RemotePath.Elem, elem)
		}
	}
	out.Resolved = in.Resolved
	out.Value = in.Value
	return out
}