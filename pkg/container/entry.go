/*
Copyright 2020 Yndd.

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

package container

import "strings"

// Entry structure keeps track of the elements in a struct/list
type Entry struct {
	Next          *Container `json:"prev,omitempty"`
	Prev          *Container `json:"next,omitempty"`
	Name          string     `json:"name,omitempty"`
	Type          string     `json:"type,omitempty"`
	Enum          []string   `json:"enum,omitempty"`
	EnumString    string     `json:"enumString,omitempty"`
	Range         []int      `json:"range,omitempty"`
	Length        []int      `json:"length,omitempty"`
	Pattern       []string   `json:"pattern,omitempty"`
	PatternString string     `json:"patternString,omitempty"`
	Union         bool       `json:"union,omitempty"`
	Mandatory     bool       `json:"mandatory,omitempty"`
	Default       string     `json:"default,omitempty"`
	Key           string     `json:"key,omitempty"`
	KeyBool       bool       `json:"keyBool,omitempty"`
	NameSpace     string     `json:"namespace,omitempty"`
}

func NewEntry(n string, opts ...EntryOption) *Entry {
	if n == "ethernet-segment" {
		n = "esi"
	}
	e := &Entry{
		Name:    n,
		Next:    nil,
		Prev:    nil,
		Enum:    make([]string, 0),
		Range:   make([]int, 0),
		Length:  make([]int, 0),
		Pattern: make([]string, 0),
	}

	for _, o := range opts {
		o(e)
	}

	return e
}

func (e *Entry) GetNext() *Container {
	return e.Next
}

func (e *Entry) GetPrev() *Container {
	return e.Prev
}

func (e *Entry) GetName() string {
	return e.Name
}

func (e *Entry) GetType() string {
	return e.Type
}

func (e *Entry) GetEnum() []string {
	return e.Enum
}

func (e *Entry) GetEnumString() string {
	return e.EnumString
}

func (e *Entry) GetRange() []int {
	return e.Range
}

func (e *Entry) GetLength() []int {
	return e.Length
}

func (e *Entry) GetPattern() []string {
	return e.Pattern
}

func (e *Entry) GetPatternString() string {
	return e.PatternString
}

func (e *Entry) GetUnion() bool {
	return e.Union
}

func (e *Entry) GetMandatory() bool {
	return e.Mandatory
}

func (e *Entry) GetDefault() string {
	return e.Default
}

func (e *Entry) GetKey() []string {
	if e.Key == "" {
		return nil
	}
	return strings.Split(e.Key, " ")
}

func (e *Entry) GetKeyBool() bool {
	return e.KeyBool
}

func (e *Entry) GetNamespace() string {
	return e.NameSpace
}