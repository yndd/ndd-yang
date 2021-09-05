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
	"github.com/yndd/ndd-runtime/pkg/logging"
)

type Parser struct {
	// logging
	log logging.Logger
}

// Option can be used to manipulate Options.
type Option func(p *Parser)

// WithLogger specifies how the Parser should log messages.
func WithLogger(log logging.Logger) Option {
	return func(p *Parser) {
		p.log = log
	}
}
func NewParser(opts ...Option) *Parser {
	p := &Parser{}

	for _, o := range opts {
		o(p)
	}

	return p
}
