/*
Copyright 2020 Wim Henderickx.

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
	"reflect"
	"testing"

	"github.com/openconfig/gnmi/proto/gnmi"
)

func TestGnmiPathToXPath(t *testing.T) {
	tests := []struct {
		inp *gnmi.Path
		exp string
	}{
		{
			inp: &gnmi.Path{
				Elem: []*gnmi.PathElem{},
			},
			exp: "/",
		},
		{
			inp: &gnmi.Path{
				Elem: []*gnmi.PathElem{
					{Name: "a"},
					{Name: "b"},
				},
			},
			exp: "/a/b",
		},
		{
			inp: &gnmi.Path{
				Elem: []*gnmi.PathElem{{
					Name: "a", Key: map[string]string{"z1": "z2"},
				}, {
					Name: "b",
				}},
			},
			exp: "/a[z1=z2]/b",
		},
		{
			inp: &gnmi.Path{
				Elem: []*gnmi.PathElem{{
					Name: "a", Key: map[string]string{"z1": "z2", "z3": "z4"},
				}, {
					Name: "b",
				}},
			},
			exp: "/a[z1=z2][z3=z4]/b",
		},
	}

	for _, tt := range tests {
		parser := NewParser()
		ret := parser.GnmiPathToXPath(tt.inp, true)
		if !reflect.DeepEqual(*ret, tt.exp) {
			t.Errorf("sortedVals(%v) = got %v, want %v", tt.inp, *ret, tt.exp)
		}
	}
}

func TestXpathToGnmiPath(t *testing.T) {
	tests := []struct {
		inp string
		exp *gnmi.Path
	}{
		{
			inp: "",
			exp: &gnmi.Path{
				Elem: []*gnmi.PathElem{},
			},
		},
		{
			inp: "/",
			exp: &gnmi.Path{
				Elem: []*gnmi.PathElem{},
			},
		},
		{
			inp: "/a/b",
			exp: &gnmi.Path{
				Elem: []*gnmi.PathElem{
					{Name: "a"},
					{Name: "b"},
				},
			},
		},
		{
			inp: "/a[z1=z2]/b",
			exp: &gnmi.Path{
				Elem: []*gnmi.PathElem{{
					Name: "a", Key: map[string]string{"z1": "z2"},
				}, {
					Name: "b",
				}},
			},
		},
		{
			inp: "/a[z1=z2, z3=z4]/b",
			exp: &gnmi.Path{
				Elem: []*gnmi.PathElem{{
					Name: "a", Key: map[string]string{"z1": "z2", "z3": "z4"},
				}, {
					Name: "b",
				}},
			},
		},
	}

	for _, tt := range tests {
		parser := NewParser()
		ret := parser.XpathToGnmiPath(tt.inp, 0)
		if !reflect.DeepEqual(ret, tt.exp) {
			t.Errorf("sortedVals(%v) = got %v, want %v", tt.inp, ret, tt.exp)
		}
	}
}
