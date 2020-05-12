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

package ltl

// A collection of useful functions for working with LTL types.

// Match is a nil-safe equivalent to op.Match().  If op is nil, NotMatching is
// returned.
func Match(op Operator, tok Token) (Operator, Environment) {
	if op != nil {
		return op.Match(tok)
	}
	return nil, NotMatching
}

// IsErroring returns true if the provided Environment's state is Erroring.
func IsErroring(e Environment) bool {
	return e.Err() != nil
}

// EitherErroring returns nil if neither of the provided Environments is
// Erroring.  Otherwise, it returns one of the Erroring arguments.
func EitherErroring(a, b Environment) Environment {
	if IsErroring(a) {
		return a
	}
	if IsErroring(b) {
		return b
	}
	return nil
}

// A nil-safe replacement for op.Reducible().  nil Operators are always
// Reducible.
func Reducible(op Operator) bool {
	if op == nil {
		return true
	}
	return op.Reducible()
}

// Reduce attempts to reduce two Environments to one.  If no reduction is
// possible, returns nil.  An Environment may be reducedif it is Reducible and
// its match status is equal to the one provided.
func Reduce(left, right Environment, matching bool) Environment {
	if left.Reducible() && (matching == left.Matching()) {
		return right
	}
	if right.Reducible() && (matching == right.Matching()) {
		return left
	}
	return nil
}
