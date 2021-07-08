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

package bindingenvironment

import (
	"fmt"
	"ltl/pkg/bindings"
	"ltl/pkg/ltl"
	"sort"
	"strings"
)

// bindingNode is the base type for bindingEnvironments.  A bare bindingNode
// represents a set of bound key-value pairs.  A bindingNode with no bound
// values is immediately simplified to a simple Matching Environment.
type bindingNode struct {
	matching   bool
	captured   map[ltl.Token]struct{}
	bound      *bindings.Bindings
	referenced *bindings.Bindings
}

// Option is used to build new bindingEnvironments.
type Option func(bn *bindingNode)

// Matching sets whether the bindingEnvironment is matching.  Defaults to true.
func Matching(m bool) Option {
	return func(bn *bindingNode) {
		bn.matching = m
	}
}

// Captured sets the bindingEnvironment's captured tokens.
func Captured(toks ...ltl.Token) Option {
	return func(bn *bindingNode) {
		bn.captured = map[ltl.Token]struct{}{}
		for _, tok := range toks {
			bn.captured[tok] = struct{}{}
		}
	}
}

// Bound sets the bindingEnvironment's bindings.  Defaults to no bindings.
func Bound(b *bindings.Bindings) Option {
	bp := &b
	return func(bn *bindingNode) {
		bn.bound = *bp
	}
}

// Bound sets the bindingEnvironment's references.  Defaults to no references.
func Referenced(r *bindings.Bindings) Option {
	rp := &r
	return func(bn *bindingNode) {
		bn.referenced = *rp
	}
}

// New returns a new bindingNode with the requested arguments applied.
// By default, the returned bindingNode is matching, and has no bound or
// referenced values.
func New(opts ...Option) *bindingNode {
	ret := &bindingNode{
		matching:   true,
		bound:      nil,
		referenced: nil,
	}
	for _, o := range opts {
		o(ret)
	}
	return ret
}

func (bn *bindingNode) String() string {
	var ret []string
	ret = append(ret, fmt.Sprintf("(%s/%t)", ltl.State(bn.Matching()), bn.matching))
	if bn.bound.Length() > 0 {
		ret = append(ret, fmt.Sprintf("BIND(%s)", bn.bound))
	}
	if bn.referenced.Length() > 0 {
		ret = append(ret, fmt.Sprintf("REF(%s)", bn.referenced))
	}
	if bn.captured != nil && len(bn.captured) > 0 {
		capStrs := []string{}
		for cap := range bn.captured {
			capStrs = append(capStrs, cap.String())
		}
		sort.Slice(capStrs, func(a, b int) bool {
			return capStrs[a] < capStrs[b]
		})
		ret = append(ret, fmt.Sprintf("CAP(%s)", strings.Join(capStrs, ", ")))
	}
	return fmt.Sprintf("(%s)", strings.Join(ret, ", "))
}

func (bn *bindingNode) And(oe ltl.Environment) ltl.Environment {
	return and(bn, oe)
}

func (bn *bindingNode) Or(oe ltl.Environment) ltl.Environment {
	return or(bn, oe)
}

func (bn *bindingNode) Not() ltl.Environment {
	// Here and elsewhere, we avoid Options to avoid allocating a jillion
	// closure functions in the critical path.
	n := New()
	n.matching = !bn.matching
	n.bound = bn.bound
	n.referenced = bn.referenced
	n.captured = bn.captured
	return n
}

func (bn *bindingNode) Matching() bool {
	if bn.hasReferences() {
		return false
	}
	return bn.matching
}

func (bn *bindingNode) Err() error {
	return nil
}

func (bn *bindingNode) Reducible() bool {
	return bn.bound.Length() == 0 &&
		bn.referenced.Length() == 0 &&
		len(bn.captured) == 0
}

func (bn *bindingNode) captures() map[ltl.Token]struct{} {
	return bn.captured
}

func (bn *bindingNode) bindings() *bindings.Bindings {
	if bn.Matching() {
		return bn.bound
	}
	return nil
}

func (bn *bindingNode) hasReferences() bool {
	return bn.referenced.Length() > 0
}

// applyBindings applies the provided Bindings to the receiver.  This returns
// a new bindingNode with:
//  * its bound field set to the receiver's bound field combinec with the
//    provided Bindings;
//  * its referenced field set to the receiver's referenced field satisfied with
//    the provided Bindings;
//  * its matching field set to:
//    * the receiver's matching field if the receiver has no references or the
//      receiver's references were satisfied by the provided Bindings;
//    * the inverse of the receiver's matching field if the receiver's
//      references were not satisfied by the provided Bindings.
// For performance, where the returned value is identical to the receiver, the
// receiver itself is returned.
// applyBindings must return an ltl.Environment, as it could return an ErrEnv.
func (bn *bindingNode) applyBindings(b *bindings.Bindings) ltl.Environment {
	if b.Length() == 0 {
		return bn
	}
	newB, err := bn.bound.Combine(b)
	if err != nil {
		return ltl.ErrEnv(err)
	}
	if !bn.hasReferences() {
		// Performance: if there's no difference in the bindings, just return
		// the receiver.
		if bn.bound.Eq(newB) {
			return bn
		}
		// If there's no references, we can simply combine bindings and return.
		new := New()
		new.captured = bn.captured
		new.matching = bn.matching
		new.bound = newB
		return new
	}
	// Otherwise, we must satisfy references.
	newR, satisfied := bn.referenced.Satisfy(newB)
	s := bn.matching
	if !satisfied {
		newR = nil
		s = !s
	}
	new := New()
	new.captured = bn.captured
	new.matching = s
	new.bound = newB
	new.referenced = newR
	return new
}

func (bn *bindingNode) merge(oe ltl.Environment) (bindingEnvironment, bool) {
	if obn, ok := oe.(*bindingNode); ok {
		if bn.matching == obn.matching &&
			bn.bound.Eq(obn.bound) &&
			bn.referenced.Eq(obn.referenced) {
			new := New()
			new.captured = UnionCaps(bn.captured, obn.captured)
			new.matching = bn.matching
			new.bound = bn.bound
			new.referenced = bn.referenced
			return new, true
		}
	}
	return nil, false
}
