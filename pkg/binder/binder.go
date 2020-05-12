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
	be "ltl/pkg/bindingenvironment"
	"ltl/pkg/bindings"
	"ltl/pkg/ltl"
	"ltl/pkg/tags"
)

// extractFunc extracts the bindings and tags from a token.
type extractFunc func(name string, tok ltl.Token) (*bindings.Bindings, *tags.Tags, error)

type binder struct {
	name         string
	extractToken extractFunc
}

func (b *binder) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	if tok.EOI() {
		return nil, be.New(be.Matching(false))
	}
	bs, tags, err := b.extractToken(b.name, tok)
	if err != nil {
		return nil, ltl.ErrEnv(err)
	}
	return nil, be.New(
		be.Tagged(tags),
		be.Bound(bs))
}

func (b *binder) String() string {
	return fmt.Sprintf("[$%s<-]", b.name)
}

func (b *binder) Reducible() bool {
	return false
}

type referencer binder

func (r *referencer) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	if tok.EOI() {
		return nil, ltl.NotMatching
	}
	bs, tags, err := r.extractToken(r.name, tok)
	if err != nil {
		return nil, ltl.ErrEnv(err)
	}
	return nil, be.New(
		be.Tagged(tags),
		be.Referenced(bs))
}

func (r *referencer) String() string {
	return fmt.Sprintf("[$%s]", r.name)
}

func (r *referencer) Reducible() bool {
	return false
}

// Builder provides methods to generate binding and referencing Operators.
type Builder struct {
	extractToken extractFunc
}

// NewBuilder returns a Builder that uses the provided extraction function to
// generate binding and referencing Operators.
func NewBuilder(extractToken func(name string, tok ltl.Token) (*bindings.Bindings, *tags.Tags, error)) *Builder {
	return &Builder{
		extractToken: extractToken,
	}
}

// Bind returns an Operator which, on Match, applies the receiver's extraction
// function to the Token to extract its bindings and tags, returning a matching
// Environment with those bindings and tags.
func (bb *Builder) Bind(name string) *binder {
	return &binder{name: name, extractToken: bb.extractToken}
}

// Reference returns an Operator which, on Match, applies the receiver's
// extraction function to the Token to extract its bindings and tags, returning
// a non-matching Environment with those tags, and referencing those bindings.
func (bb *Builder) Reference(name string) *referencer {
	return &referencer{name: name, extractToken: bb.extractToken}
}
