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

// Package ltl defines a set of core types and functions for framing and
// applying linear temporal logic (LTL) queries.
package ltl

import (
	"fmt"
)

// Token represents an input token to a query.  nil Tokens are not valid.
// Tokens, once created, should not be modified.
type Token interface {
	fmt.Stringer
	// EOI should return true if the receiver marks the end of an input stream.
	EOI() bool
}

// Environment represents the environment of a query.  nil Environments are not
// valid.  Environments, once created, should not be modified.
type Environment interface {
	fmt.Stringer
	// And returns the logical AND of the receiver and provided Environment,
	// including their matching states, and where applicable, any other state
	// they contain.
	// And must not modify either its receiver or its argument.  It may return
	// its receiver or argument, if semantically appropriate.
	And(env Environment) Environment
	// Or returns the logical OR of the receiver and provided Environment,
	// including their matching states, and where applicable, any other state
	// they contain.
	// Or must not modify either its receiver or its argument.  It may return
	// its reveiver or argument, if semantically appropriate.
	Or(env Environment) Environment
	// Not returns the logical NOT of the receiver, including any required
	// changes to stored state.
	// Not must not modify its receiver.
	Not() Environment
	// Matching should return true iff the receiver is matching.
	Matching() bool
	// Err should return the Environment's error, which may be nil.
	Err() error
	// Reducible should return true iff this Environment's only state is its
	// Matching and Err status.  Irreducible Environments convey sideband state
	// that cannot safely be discarded, or their final Matching state is
	// pending receiving such sideband state.
	// It is always safe to return false, but this may impact performance.
	Reducible() bool
}

// Operator represents a LTL query operator.  A nil Operator should be
// construed as always returning NotMatching on a Match.
type Operator interface {
	// If the receiver intends to support prettyprint.PrettyPrint, String()
	// should return only the name and state of the receiver, but not its
	// children.
	fmt.Stringer
	// Match applies a Token to the receiving Operation.  It should return the
	// Environment of the query after the token is applied.  It also returns an
	// Operator reflecting the continuation of the query after the token is
	// consumed.  This Operator may be nil, indicating that no further tokens
	// may be consumed by this query.  An error encountered while processing the
	// query should be indicated by returning an Erroring environment.
	Match(Token) (Operator, Environment)
	// Reducible should return true iff this Operator's Match function can
	// *only* return Reducible Environments -- if it might return an irreducible
	// Environment, Reducible must return false.
	// It is always safe to return false, but this may impact performance.
	Reducible() bool
}
