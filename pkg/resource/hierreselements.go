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

package resource

import (
	"fmt"
	"strings"
)

type HierResourceElements struct {
	Elems map[string]interface{}
}

func NewHierResourceElements() *HierResourceElements {
	return &HierResourceElements{
		Elems: make(map[string]interface{}),
	}
}

// builds a hierarchical map[string]map[string]nil element
func (h *HierResourceElements) GetHierResourceElements() map[string]interface{} {
	return h.Elems
}

// builds a hierarchical map[string]map[string]nil element
func (h *HierResourceElements) AddHierResourceElement(path string) {
	h.Elems = addHierResourceElement(h.Elems, strings.Split(path, "/")[1:])
}

func addHierResourceElement(h map[string]interface{}, e []string) map[string]interface{} {
	fmt.Printf("addHierResourceElement: %v\n", e)
	if len(e) > 1 {
		// not last element
		// check if it was already initialized
		if _, ok := h[e[0]]; !ok {
			h[e[0]] = make(map[string]interface{})
		}
		switch x := h[e[0]].(type) {
		case map[string]interface{}:
			h[e[0]] = addHierResourceElement(x, e[1:])
		}
	} else {
		// last elment
		h[e[0]] = nil
	}
	return h
}
