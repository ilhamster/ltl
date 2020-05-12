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

package operators

import (
	"ltl/pkg/ltl"
)

// A collection of base types for non-terminal ltl.Operator implementations.
// New ltl.Operator definitions can embed the appropriate type for PrettyPrint
// support.

// UnaryOperator is a base type for ltl.Operators with one child ltl.Operator.
type UnaryOperator struct {
	Child ltl.Operator
}

func (uo UnaryOperator) Children() []ltl.Operator {
	return []ltl.Operator{uo.Child}
}

func (uo UnaryOperator) Reducible() bool {
	return uo.Child.Reducible()
}

// BinaryOperator is a base type for ltl.Operators with two child ltl.Operators.
type BinaryOperator struct {
	Left, Right ltl.Operator
}

func (bo BinaryOperator) Children() []ltl.Operator {
	return []ltl.Operator{bo.Left, bo.Right}
}

func (bo BinaryOperator) Reducible() bool {
	return bo.Left.Reducible() || bo.Right.Reducible()
}

// MatchBoth applies the provided ltl.Token to both child ltl.Operators of the
// receiver.
func (bo BinaryOperator) MatchBoth(tok ltl.Token) (newLeft, newRight ltl.Operator, leftEnv, rightEnv ltl.Environment) {
	newLeft, leftEnv = ltl.Match(bo.Left, tok)
	newRight, rightEnv = ltl.Match(bo.Right, tok)
	return
}

// NaryOperator is a base type for ltl.Operators with arbitrary child
// ltl.Operators.
type NaryOperator struct {
	ChildSlice []ltl.Operator
}

func (no NaryOperator) Children() []ltl.Operator {
	return no.ChildSlice
}

func (no NaryOperator) Reducible() bool {
	for _, child := range no.ChildSlice {
		if child.Reducible() {
			return true
		}
	}
	return false
}
