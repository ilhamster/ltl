// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package binder provides a type, BinderBuilder, for easy construction of
// binding and referencing Operators.  See docs/binding.md.
package binder

import (
	"fmt"
	be "github.com/ilhamster/ltl/pkg/bindingenvironment"
	"github.com/ilhamster/ltl/pkg/bindings"
	"github.com/ilhamster/ltl/pkg/ltl"
)

// extractFunc extracts the bindings and tags from a token.
type extractFunc func(name string, tok ltl.Token) (*bindings.Bindings, error)

// Binder is an Operator capable of binding values from tokens.  A bound value
// satisfies other bound and referenced instances of the same value.
type Binder struct {
	name         string
	capture      bool
	extractToken extractFunc
}

// Match performs an LTL match on the receiving Binder.
func (b *Binder) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	if tok.EOI() {
		return nil, be.New(be.Matching(false))
	}
	bs, err := b.extractToken(b.name, tok)
	if err != nil {
		return nil, ltl.ErrEnv(err)
	}
	ops := []be.Option{be.Bound(bs)}
	if b.capture {
		ops = append(ops, be.Captured(tok))
	}
	return nil, be.New(ops...)
}

func (b *Binder) String() string {
	return fmt.Sprintf("[$%s<-]", b.name)
}

// Reducible returns false for all Binders.
func (b *Binder) Reducible() bool {
	return false
}

// Referencer is an Operator capable of referencing values from tokens.  A
// referenced value is satisfied by a bound instance of the same value.
type Referencer Binder

// Match performs an LTL match on the receiving Referencer.
func (r *Referencer) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	if tok.EOI() {
		return nil, ltl.NotMatching
	}
	bs, err := r.extractToken(r.name, tok)
	if err != nil {
		return nil, ltl.ErrEnv(err)
	}
	ops := []be.Option{be.Referenced(bs)}
	if r.capture {
		ops = append(ops, be.Captured(tok))
	}
	return nil, be.New(ops...)
}

func (r *Referencer) String() string {
	return fmt.Sprintf("[$%s]", r.name)
}

// Reducible returns false for all Referencers.
func (r *Referencer) Reducible() bool {
	return false
}

// Builder provides methods to generate binding and referencing Operators.
type Builder struct {
	extractToken extractFunc
	capture      bool
}

// NewBuilder returns a Builder that uses the provided extraction function to
// generate binding and referencing Operators.
func NewBuilder(capture bool, extractToken func(name string, tok ltl.Token) (*bindings.Bindings, error)) *Builder {
	return &Builder{
		extractToken: extractToken,
		capture:      capture,
	}
}

// Bind returns an Operator which, on Match, applies the receiver's extraction
// function to the Token to extract its bindings, returning a matching
// Environment with those bindings.
func (bb *Builder) Bind(name string) *Binder {
	return &Binder{name: name, capture: bb.capture, extractToken: bb.extractToken}
}

// Reference returns an Operator which, on Match, applies the receiver's
// extraction function to the Token to extract its bindings, returning a
// non-matching Environment with those, and referencing those bindings.
func (bb *Builder) Reference(name string) *Referencer {
	return &Referencer{name: name, capture: bb.capture, extractToken: bb.extractToken}
}
