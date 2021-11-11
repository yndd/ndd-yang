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

import (
	"github.com/openconfig/gnmi/proto/gnmi"
)

type LeafRefValidationKind string

const (
	LeafRefValidationLocal    LeafRefValidationKind = "local"
	LeafRefValidationExternal LeafRefValidationKind = "external"
)

type LeafRef struct {
	LocalPath  *gnmi.Path `json:"localPath,omitempty"`
	RemotePath *gnmi.Path `json:"remotePath,omitempty"`
}
