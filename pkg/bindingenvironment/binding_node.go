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
	"ltl/pkg/captures"
	"ltl/pkg/ltl"
	"sort"
	"strings"
)

// BindingNode is the base type for bindingEnvironments.  A bare BindingNode
// represents a set of bound key-value pairs.  A BindingNode with no bound
// values is immediately simplified to a simple Matching Environment.
type BindingNode struct {
	matching   bool
	caps       *captures.Captures
	bound      *bindings.Bindings
	referenced *bindings.Bindings
}

// Option is used to build new bindingEnvironments.
type Option func(bn *BindingNode)

// Matching sets whether the bindingEnvironment is matching.  Defaults to true.
func Matching(m bool) Option {
	return func(bn *BindingNode) {
		if bn.matching != m {
			bn.matching = m
			bn.caps = bn.caps.Not()
		}
	}
}

// Captured sets the bindingEnvironment's captured tokens.
func Captured(toks ...ltl.Token) Option {
	return func(bn *BindingNode) {
		cap := map[ltl.Token]struct{}{}
		for _, tok := range toks {
			cap[tok] = struct{}{}
		}
		bn.caps = captures.New()
		bn.caps.Capture(bn.matching, toks...)
	}
}

// Bound sets the bindingEnvironment's bindings.  Defaults to no bindings.
func Bound(b *bindings.Bindings) Option {
	bp := &b
	return func(bn *BindingNode) {
		bn.bound = *bp
	}
}

// Referenced sets the bindingEnvironment's references.  Defaults to no
// references.
func Referenced(r *bindings.Bindings) Option {
	rp := &r
	return func(bn *BindingNode) {
		bn.referenced = *rp
	}
}

// New returns a new BindingNode with the requested arguments applied.
// By default, the returned BindingNode is matching, and has no bound or
// referenced values.
func New(opts ...Option) *BindingNode {
	ret := &BindingNode{
		matching:   true,
		bound:      nil,
		referenced: nil,
	}
	for _, o := range opts {
		o(ret)
	}
	return ret
}

func (bn *BindingNode) String() string {
	var ret []string
	ret = append(ret, fmt.Sprintf("(%s/%t)", ltl.State(bn.Matching()), bn.matching))
	if bn.bound.Length() > 0 {
		ret = append(ret, fmt.Sprintf("BIND(%s)", bn.bound))
	}
	if bn.referenced.Length() > 0 {
		ret = append(ret, fmt.Sprintf("REF(%s)", bn.referenced))
	}
	caps := bn.captures().Get(bn.matching)
	// caps := bn.matchingCaptured
	if caps != nil && len(caps) > 0 {
		capStrs := []string{}
		for cap := range caps {
			capStrs = append(capStrs, cap.String())
		}
		sort.Slice(capStrs, func(a, b int) bool {
			return capStrs[a] < capStrs[b]
		})
		ret = append(ret, fmt.Sprintf("CAP(%s)", strings.Join(capStrs, ", ")))
	}
	return fmt.Sprintf("(%s)", strings.Join(ret, ", "))
}

// And returns the AND of the receiver and argument.
func (bn *BindingNode) And(oe ltl.Environment) ltl.Environment {
	return and(bn, oe)
}

// Or returns the OR of the receiver and argument.
func (bn *BindingNode) Or(oe ltl.Environment) ltl.Environment {
	return or(bn, oe)
}

// Not returns the NOT of the receiver.
func (bn *BindingNode) Not() ltl.Environment {
	// Here and elsewhere, we avoid Options to avoid allocating a jillion
	// closure functions in the critical path.
	n := New()
	n.matching = !bn.matching
	n.bound = bn.bound
	n.referenced = bn.referenced
	n.caps = bn.caps.Not()
	return n
}

// Matching returns false for any BindingNode that has references, since these
// are still pending, and otherwise the node's matching status.
func (bn *BindingNode) Matching() bool {
	if bn.hasReferences() {
		return false
	}
	return bn.matching
}

// Err returns nil for all BindingNodes.
func (bn *BindingNode) Err() error {
	return nil
}

// Reducible returns true for BindingNodes with no bound values, references, or
// captures.
func (bn *BindingNode) Reducible() bool {
	return bn.bound.Length() == 0 &&
		bn.referenced.Length() == 0 &&
		bn.caps.Reducible()
}

func (bn *BindingNode) captures() *captures.Captures {
	return bn.caps
}

func (bn *BindingNode) bindings() *bindings.Bindings {
	if bn.Matching() {
		return bn.bound
	}
	return nil
}

func (bn *BindingNode) hasReferences() bool {
	return bn.referenced.Length() > 0
}

// applyBindings applies the provided Bindings to the receiver.  This returns
// a new BindingNode with:
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
func (bn *BindingNode) applyBindings(b *bindings.Bindings) ltl.Environment {
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
		new.caps = bn.caps
		new.matching = bn.matching
		new.bound = newB
		return new
	}
	new := New()
	new.caps = bn.caps
	new.matching = bn.matching
	// Otherwise, we must satisfy references.
	newR, satisfied := bn.referenced.Satisfy(newB)
	if !satisfied {
		newR = nil
		new = new.Not().(*BindingNode)
	}
	new.bound = newB
	new.referenced = newR
	return new
}

func (bn *BindingNode) merge(oe ltl.Environment) (bindingEnvironment, bool) {
	if obn, ok := oe.(*BindingNode); ok {
		if bn.matching == obn.matching &&
			bn.bound.Eq(obn.bound) &&
			bn.referenced.Eq(obn.referenced) {
			new := New()
			new.caps = bn.caps.Union(obn.caps)
			new.matching = bn.matching
			new.bound = bn.bound
			new.referenced = bn.referenced
			return new, true
		}
	}
	return nil, false
}
