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
	"fmt"
	"strconv"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/stoewer/go-strcase"
	"github.com/yndd/ndd-yang/pkg/container"
)

// GetypeName return a string of the type of the  yang entry
func (p *Parser) GetTypeName(e *yang.Entry) string {
	if e == nil || e.Type == nil {
		return ""
	}
	// Return our root's type name.
	// This is should be the builtin type-name
	// for this entry.
	return e.Type.Name
}

// GetTypeKind return a string of the kind of the yang entry
func (p *Parser) GetTypeKind(e *yang.Entry) string {
	if e == nil || e.Type == nil {
		return ""
	}
	// Return our root's type name.
	// This is should be the builtin type-name
	// for this entry.
	return e.Type.Kind.String()
}

// CreatePathElem returns a config path element from a yang Entry
func (p *Parser) CreatePathElem(e *yang.Entry) *gnmi.PathElem {
	pathElem := &gnmi.PathElem{
		Name: e.Name,
		Key:  make(map[string]string),
	}

	if e.Key != "" {
		var keyType string
		switch p.GetTypeName(e.Dir[e.Key]) {
		case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
			keyType = p.GetTypeName(e.Dir[e.Key])
		case "boolean":
			keyType = "bool"
		case "enumeration":
			keyType = "string"
		default:
			keyType = "string"
		}
		pathElem.Key[e.Key] = keyType
		//fmt.Printf("Key: %s, KeyType: %s\n", e.Key, keyType)
	}
	return pathElem
}

func (p *Parser) CreateContainerEntry(e *yang.Entry, next, prev *container.Container, containerKey string) *container.Entry {
	// Allocate a new Entry
	entry := container.NewEntry(e.Name)

	// initialize the Next pointer if relevant -> only relevant for list
	entry.Next = next
	entry.Prev = prev

	entry.NameSpace = e.Namespace().Name

	if e.Name == "port-binding" {
		fmt.Printf("port-binding: choice: %#v Identities: %#v, Other: %#v\n", e.IsChoice(), e.Identities, e.Exts)

	}

	

	// process mandatory attribute
	switch e.Mandatory {
	case 1: // TSTrue
		entry.Mandatory = true
	default: // TSTrue
		entry.Mandatory = false
	}
	if e.Key != "" {
		entry.Mandatory = true
	}
	// a containerkey can consists of multiple keys.
	containerKeys := strings.Split(containerKey, " ")
	// keys come from the previous container so we need to check the elements against these key(s)
	for _, containerKey := range containerKeys {
		if e.Name == containerKey {
			entry.Mandatory = true
		}
	}
	
	// process type attribute
	switch p.GetTypeName(e) {
	case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
		entry.Type = p.GetTypeName(e)
	case "boolean":
		entry.Type = "bool"
	case "enumeration":
		entry.Enum = e.Type.Enum.Names()
		entry.Type = "string"
	default:
		switch p.GetTypeKind(e) {
		case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
			entry.Type = p.GetTypeKind(e)
		case "boolean":
			entry.Type = "bool"
		case "union":
			entry.Type = "string"
			entry.Union = true
			for _, t := range e.Type.Type {
				entry.Type = t.Root.Kind.String()
				if entry.Type == "enumeration" ||
					entry.Type == "leafref" ||
					entry.Type == "union" {
					entry.Type = "string"
				}
				entry.Pattern = append(entry.Pattern, t.Pattern...)

			}
		case "leafref":
			// The processing of leaf refs is handled in another function
			entry.Type = "string"
		default:
			entry.Type = "string"
		}
	}
	// process elementType for a Key
	if e.Key != "" {
		switch p.GetTypeName(e.Dir[e.Key]) {
		case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
			entry.Type = p.GetTypeName(e.Dir[e.Key])
		case "boolean":
			entry.Type = "bool"
		default:
			entry.Type = "string"
		}
	}
	// update the Type to reflect the reference to the proper struct
	if entry.Prev != nil {
		entry.Type = strcase.UpperCamelCase(entry.Prev.GetFullName() + "-" + e.Name)
	}

	if e.Type != nil {
		for _, ra := range e.Type.Range {
			entry.Range = append(entry.Range, int(ra.Min.Value))
			// this is to account for the fact that range can be defined as 1..max
			if ra.Max.Value < ra.Min.Value {
				switch {
				case strings.Contains(entry.Type, "8"):
					entry.Range = append(entry.Range, int(255))
				case strings.Contains(entry.Type, "16"):
					entry.Range = append(entry.Range, int(65535))
				default:
					entry.Range = append(entry.Range, int(4294967295))
				}
			} else {
				entry.Range = append(entry.Range, int(ra.Max.Value))
			}
			//fmt.Printf("RANGE MIN: %d MAX: %d, TOTAL: %d\n", ra.Min.Value, ra.Max.Value, entry.Range)
		}

		for _, le := range e.Type.Length {
			entry.Length = append(entry.Length, int(le.Min.Value))
			entry.Length = append(entry.Length, int(le.Max.Value))
			//fmt.Printf("LENGTH MIN: %d MAX: %d, TOTAL: %d\n", le.Min.Value, le.Max.Value, entry.Length)
		}

		if e.Type.Pattern != nil {
			entry.Pattern = append(entry.Pattern, e.Type.Pattern...)
			//fmt.Printf("LEAF NAME: %s, PATTERN: %s\n", e.Name, entry.Pattern)

		}
		if e.Type.Kind.String() == "enumeration" {
			entry.Enum = e.Type.Enum.Names()
		}
		// DEFAULTS ARE NOT USED In PROVIDERS SINCE IT CREATES LOTS OF DEPENDENCIES BECAUSE RESOURCES HAVE OTHER DEPENDENCIES
		// AND CONTEXT. E.g. allow-icmp-redirect in sros is only supported in management context; gre-termination in a primary
		// interface is not supported in all circumstances in sros, etc etc
		// SEEMS BETER TO NOT USE UT WITH PROVIDERS
		if e.Default != "" {
			// if there is a default parameter and the entry type is a int, we will try to convert
			// it and if it does not work we dont initialize the default
			switch {
			case strings.Contains(entry.Type, "int"):
				// e.g. we can have rdnss-lifetime which has a default of infinite but it is an int32
				if _, err := strconv.Atoi(e.Default); err == nil {
					entry.Default = e.Default
				}
				// if the conversion does not succeed we dont initialize a default
			default:
				entry.Default = e.Default
			}
			fmt.Printf("Default: Type: %s, Default: %s\n", entry.Type, entry.Default)
		}

	}

	// pattern post processing
	var pattern string
	for i, p := range entry.Pattern {
		//fmt.Printf("Pattern: %s last\n", p)
		if i == (len(entry.Pattern) - 1) {
			pattern += p
		} else {
			pattern += p + "|"
		}
	}
	if len(pattern) > 0 {
		//fmt.Printf("Pattern orig: %sorig\n", pattern)
		//pattern = strings.ReplaceAll(pattern, "@", "")
		//pattern = strings.ReplaceAll(pattern, "#", "")
		//pattern = strings.ReplaceAll(pattern, "$", "")
		entry.PatternString = strings.ReplaceAll(pattern, "%", "")

		if strings.Contains(pattern, "`") {
			entry.PatternString = fmt.Sprintf("\"%s\"", entry.PatternString)
		} else {
			entry.PatternString = fmt.Sprintf("`%s`", entry.PatternString)
		}
		//fmt.Printf("Pattern processed: %sprocessed\n", pattern)
	}

	// enum post processing
	for _, enum := range entry.Enum {
		entry.EnumString += "`" + enum + "`;"
	}
	if entry.EnumString != "" {
		entry.EnumString = strings.TrimRight(entry.EnumString, ";")
	}

	// key handling
	entry.Key = e.Key

	/*
		if entry.Mandatory {
			fmt.Printf("entry.Name: %s, entry.Key: %s, e.Mandatory: %t\n", entry.Name, entry.Key, entry.Mandatory)
		}
		if entry.Key != "" {
			fmt.Printf("entry.Name: %s, entry.Key: %s, e.Mandatory: %t\n", entry.Name, entry.Key, entry.Mandatory)
		}
		if entry.Name == "router-name" {
			fmt.Printf("entry.Name: %s, entry.Key: %s, e.Mandatory: %t\n", entry.Name, entry.Key, entry.Mandatory)
		}
	*/
	return entry
}
