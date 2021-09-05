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
	"github.com/openconfig/gnmi/proto/gnmi"
)

const (
	// values
	Slice    = "slice"
	NonSlice = "nonSlice"
	// errors
	ErrJSONMarshal       = "cannot marshal JSON object"
	ErrJSONCompare       = "cannot compare JSON objects"
	ErrJSONMarshalIndent = "cannot marshal JSON object with indent"
)

// ConfigTreeAction defines the states the resource object is reporting
type ConfigTreeAction string

const (
	ConfigTreeActionGet    ConfigTreeAction = "get"
	ConfigTreeActionDelete ConfigTreeAction = "delete"
	ConfigTreeActionCreate ConfigTreeAction = "create"
	ConfigTreeActionUpdate ConfigTreeAction = "update"
	ConfigTreeActionFind   ConfigTreeAction = "find"
	ConfigResolveLeafRef   ConfigTreeAction = "resolve leafref"
)

func (c *ConfigTreeAction) String() string {
	switch *c {
	case ConfigTreeActionGet:
		return "get"
	case ConfigTreeActionDelete:
		return "delete"
	case ConfigTreeActionCreate:
		return "create"
	case ConfigTreeActionUpdate:
		return "update"
	case ConfigTreeActionFind:
		return "find"
	case ConfigResolveLeafRef:
		return "resolve leafref"
	}
	return ""
}

type TraceCtxtGnmi struct {
	Action           ConfigTreeAction
	Found            bool
	Idx              int
	Path             *gnmi.Path             // the input path data
	ResolvedLeafRefs []*ResolvedLeafRefGnmi // holds all the resolved leafRefs if they get resolved
	Data             interface{}
	Value            interface{} // used for leafref resolution
	Msg              []string
}

func (tc *TraceCtxtGnmi) AddMsg(s string) {
	tc.Msg = append(tc.Msg, s)
}
