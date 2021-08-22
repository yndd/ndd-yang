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

	config "github.com/netw-device-driver/ndd-grpc/config/configpb"
	"github.com/openconfig/goyang/pkg/yang"
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
func (p *Parser) CreatePathElem(e *yang.Entry) *config.PathElem {
	pathElem := &config.PathElem{
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
		fmt.Printf("Key: %s, KeyType: %s\n", e.Key, keyType)
	}
	return pathElem
}
