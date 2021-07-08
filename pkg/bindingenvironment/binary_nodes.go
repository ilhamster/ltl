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

type nodeType bool

const (
	andNode nodeType = false
	orNode  nodeType = true
)

type binaryNode struct {
	bound       *bindings.Bindings
	left, right ltl.Environment
	hasRefs     bool
	matching    bool
	t           nodeType
}

func (bn *binaryNode) String() string {
	var ret string
	switch bn.t {
	case orNode:
		ret = "BE_OR"
	case andNode:
		ret = "BE_AND"
	}
	capStrs := []string{}
	caps := bn.captures()
	if caps != nil {
		for cap := range caps {
			capStrs = append(capStrs, cap.String())
		}
		sort.Slice(capStrs, func(a, b int) bool {
			return capStrs[a] < capStrs[b]
		})
	}
	return ret + fmt.Sprintf("(r:%t\n    | M%t/%s,\n     |C[%s] | %s,\n    | %s)", bn.hasRefs, bn.Matching(), bn.bound, strings.Join(capStrs, ", "), bn.left, bn.right)
}

func (bn *binaryNode) And(oe ltl.Environment) ltl.Environment {
	return and(bn, oe)
}

func (bn *binaryNode) Or(oe ltl.Environment) ltl.Environment {
	return or(bn, oe)
}

func (bn *binaryNode) Not() ltl.Environment {
	// To avoid introducing a notNode type, we use DeMorgan's laws.
	switch bn.t {
	case orNode:
		//   NOT OR(a, b)
		// : NOT (NOT AND(NOT(a), NOT(b))
		// : AND(NOT(a), NOT(b))
		return and(bn.left.Not(), bn.right.Not())
	case andNode:
		//   NOT AND(a, b)
		// : NOT (NOT OR(NOT(a), NOT(b))
		// : OR(NOT(a), NOT(b))
		return or(bn.left.Not(), bn.right.Not())
	}
	return ltl.ErrEnv(fmt.Errorf("unknown binaryNode type %v", bn.t))
}

func (bn *binaryNode) Reducible() bool {
	return false
}

func (bn *binaryNode) Matching() bool {
	return bn.matching
}

func (bn *binaryNode) Err() error {
	return nil
}

func (bn *binaryNode) captures() map[ltl.Token]struct{} {
	var left, right map[ltl.Token]struct{}
	if bn.left.Matching() {
		left = Captures(bn.left)
	}
	if bn.right.Matching() {
		right = Captures(bn.right)
	}
	return UnionCaps(left, right)
}

func (bn *binaryNode) bindings() *bindings.Bindings {
	return bn.bound
}

func (bn *binaryNode) hasReferences() bool {
	return bn.hasRefs
}

func (bn *binaryNode) applyBindings(b *bindings.Bindings) ltl.Environment {
	switch bn.t {
	case orNode:
		return or(applyBindings(b, bn.left), applyBindings(b, bn.right))
	case andNode:
		return and(applyBindings(b, bn.left), applyBindings(b, bn.right))
	}
	return ltl.ErrEnv(fmt.Errorf("unknown binaryNode type %v", bn.t))
}

func (bn *binaryNode) merge(oe ltl.Environment) (bindingEnvironment, bool) {
	// a non-binaryNode cannot be equal to a binaryNode.
	obn, ok := oe.(*binaryNode)
	if !ok {
		return nil, false
	}
	// If the rolled-up properties are not equal, the two are not equal.
	if bn.t != obn.t ||
		bn.matching != obn.matching ||
		bn.hasRefs != obn.hasRefs ||
		!bn.bound.Eq(obn.bound) {
		return nil, false
	}
	// If the children of the two binaryNodes cannot be merged pairwise, the
	// two are not equal.
	newL, newLOk := merge(bn.left, obn.left)
	newR, newROk := merge(bn.right, obn.right)
	if !newLOk || !newROk {
		newL, newLOk = merge(bn.left, obn.right)
		newR, newROk = merge(bn.right, obn.left)
		if !newLOk || !newROk {
			return nil, false
		}
	}
	return &binaryNode{
		bound:    bn.bound,
		left:     newL,
		right:    newR,
		hasRefs:  bn.hasRefs,
		matching: bn.matching,
		t:        bn.t,
	}, true
}

// and builds and returns a new andNode representing the AND of its two
// arguments.  If either argument has a non-nil Err(), it returns that instead,
// and if either argument is reducible and matching, the other argument is
// returned instead.
func and(left, right ltl.Environment) ltl.Environment {
	if errEnv := ltl.EitherErroring(left, right); errEnv != nil {
		return errEnv
	}
	if red := ltl.Reduce(left, right, true); red != nil {
		return red
	}
	newB, err := Bindings(left).Combine(Bindings(right))
	if err != nil {
		return ltl.ErrEnv(err)
	}
	left = applyBindings(newB, left)
	right = applyBindings(newB, right)
	if ret, ok := merge(left, right); ok {
		return ret
	}
	hasRefs := hasReferences(left) || hasReferences(right)
	matching := false
	if !hasRefs {
		matching = left.Matching() && right.Matching()
	}
	return &binaryNode{
		bound:    newB,
		left:     left,
		right:    right,
		hasRefs:  hasRefs,
		matching: matching,
		t:        andNode,
	}
}

// or builds and returns a new orNode representing the OR of its two arguments.
// If either argument has a non-nil Err(), it returns that instead, and if
// either argument is reducible and not matching, the other argument is returned
// instead.
func or(left, right ltl.Environment) ltl.Environment {
	if errEnv := ltl.EitherErroring(left, right); errEnv != nil {
		return errEnv
	}
	if red := ltl.Reduce(left, right, false); red != nil {
		return red
	}
	newB, err := Bindings(left).Combine(Bindings(right))
	if err != nil {
		return ltl.ErrEnv(err)
	}
	left = applyBindings(newB, left)
	right = applyBindings(newB, right)
	if ret, ok := merge(left, right); ok {
		return ret
	}
	hasRefs := hasReferences(left) || hasReferences(right)
	matching := false
	if !hasRefs {
		matching = left.Matching() || right.Matching()
	}
	return &binaryNode{
		bound:    newB,
		left:     left,
		right:    right,
		hasRefs:  hasRefs,
		matching: matching,
		t:        orNode,
	}
}
