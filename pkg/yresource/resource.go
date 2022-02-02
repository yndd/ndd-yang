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

package yresource

import (
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"github.com/yndd/ndd-yang/pkg/leafref"
	"github.com/yndd/ndd-yang/pkg/yentry"
)

type Resource struct {
	Log          logging.Logger
	DeviceSchema *yentry.Entry
}

func (r *Resource) WithLogging(log logging.Logger) {
	r.Log = log
}

type Handler interface {
	WithLogging(log logging.Logger)
	GetRootPath(mg resource.Managed) []*gnmi.Path
	GetParentDependency(mg resource.Managed) []*leafref.LeafRef
}

type Option func(Handler)

func WithLogging(log logging.Logger) Option {
	return func(o Handler) {
		o.WithLogging(log)
	}
}
