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

// Package operators defines a collection of LTL logical and temporal operators.
package operators

import (
	"fmt"
	"ltl/pkg/ltl"
)

// StopAtFirstMatch matches the provided Operator with the provided Token.
// The resulting Operation and Environment are returned, except if the
// Environment is Matching, in which case a nil Operator is returned.  This
// helps Operators terminate as soon as they've matched, a necessary property
// for temporal operators like Then to work.
func StopAtFirstMatch(tok ltl.Token, op ltl.Operator) (ltl.Operator, ltl.Environment) {
	op, env := op.Match(tok)
	if env.Matching() {
		op = nil
	}
	return op, env
}

// StopAtFirstNotMatch matches the provided Operator with the provided Token.
// The resulting Operation and Environment are returned, except if the
// Environment is not Matching, in which case a nil Operator is returned.  This
// helps Operators terminate as soon as they've matched, a necessary property
// for temporal operators like Then to work.
func StopAtFirstNotMatch(tok ltl.Token, op ltl.Operator) (ltl.Operator, ltl.Environment) {
	op, env := op.Match(tok)
	if !env.Matching() {
		op = nil
	}
	return op, env
}

// Not is the logical NOT of its argument, inverting the Environments it
// returns.
func Not(child ltl.Operator) ltl.Operator {
	if child == nil {
		return nil
	}
	return &not{UnaryOperator{child}}
}

type not struct {
	UnaryOperator
}

func (n *not) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	newOp, env := ltl.Match(n.Child, tok)
	return Not(newOp), env.Not()
}

func (n *not) String() string {
	return "NOT"
}

// And is the logical AND of its arguments.
func And(left, right ltl.Operator) ltl.Operator {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	return &and{BinaryOperator{left, right}}
}

type and struct {
	BinaryOperator
}

func (a *and) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	newLeft, newRight, leftEnv, rightEnv := a.BinaryOperator.MatchBoth(tok)
	if errEnv := ltl.EitherErroring(leftEnv, rightEnv); errEnv != nil {
		return nil, errEnv
	}
	// If one child resolves before the other, we use an AndEnvironment to
	// store it until the other side is ready.  This is required for And, since
	// both sides are necessary, but not required of Or, which simply returns
	// the first of its sides to resolve.
	newEnv := leftEnv.And(rightEnv)
	if newLeft == nil {
		return AndEnvironment(leftEnv, newRight), newEnv
	}
	if newRight == nil {
		return AndEnvironment(rightEnv, newLeft), newEnv
	}
	return And(newLeft, newRight), newEnv
}

func (a *and) String() string {
	return "AND"
}

// Or is the logical OR of its arguments.
func Or(left, right ltl.Operator) ltl.Operator {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	return &or{BinaryOperator{left, right}}
}

type or struct {
	BinaryOperator
}

func (o *or) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	newLeft, newRight, leftEnv, rightEnv := o.BinaryOperator.MatchBoth(tok)
	if errEnv := ltl.EitherErroring(leftEnv, rightEnv); errEnv != nil {
		return nil, errEnv
	}
	newEnv := leftEnv.Or(rightEnv)
	return Or(newLeft, newRight), newEnv
}

func (o *or) String() string {
	return "OR"
}

// Limit is equivalent to the provided Operator, except that if that Operator
// does not resolve within the specified number of tokens, it returns a
// non-Matching environment.
func Limit(n int64, child ltl.Operator) ltl.Operator {
	if n == 0 || child == nil {
		return nil
	}
	return &limit{UnaryOperator{child}, n}
}

type limit struct {
	UnaryOperator
	n int64
}

func (l *limit) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	if l.n == 0 {
		return nil, ltl.NotMatching
	}
	op, env := l.Child.Match(tok)
	newOp := Limit(l.n-1, op)
	return newOp, env
}

func (l *limit) String() string {
	return fmt.Sprintf("LIMIT(%d)", l.n)
}

// Next ignores a single input token then attempts to match its child.
func Next(child ltl.Operator) ltl.Operator {
	if child == nil {
		return nil
	}
	return &next{UnaryOperator{child}}
}

type next struct {
	UnaryOperator
}

func (n *next) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	return n.Child, ltl.NotMatching
}

func (n *next) String() string {
	return "NEXT"
}

// AndEnvironment defers its argument Environment for later ANDing with the
// Environments produced by matching with its child.
func AndEnvironment(env ltl.Environment, child ltl.Operator) ltl.Operator {
	if child == nil {
		return nil
	}
	// Short-circuit: if the attached environment is Matching and Reducible,
	// we can simply return the child.
	if env.Reducible() && env.Matching() {
		return child
	}
	return &andEnvironment{UnaryOperator{child}, env}
}

type andEnvironment struct {
	UnaryOperator
	env ltl.Environment
}

// Match returns non-nil Operators as long as its child does.
func (ae *andEnvironment) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	// Short-circuit: if the bundled Environment is not matching and the
	// child Operator is reducible, there's no need to recurse, since the
	// bundled Environment is all it will ever be.
	if !ae.env.Matching() && ltl.Reducible(ae.Child) {
		return nil, ae.env
	}
	newOp, newEnv := ltl.Match(ae.Child, tok)
	return AndEnvironment(ae.env, newOp), ae.env.And(newEnv)
}

func (ae *andEnvironment) String() string {
	return fmt.Sprintf("AND_ENVIRONMENT(%s)", ae.env)
}

// OrEnvironment waits until its provided Operator resolves, then returns
// the provided Environment Ored with that resolution.
func OrEnvironment(env ltl.Environment, child ltl.Operator) ltl.Operator {
	if child == nil {
		return nil
	}
	// Short-circuit: if the attached Environment is not Matching and is
	// Reducible, we can simply return the child.
	if env.Reducible() && !env.Matching() {
		return child
	}
	return &orEnvironment{UnaryOperator{child}, env}
}

type orEnvironment struct {
	UnaryOperator
	env ltl.Environment
}

func (oe *orEnvironment) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	newOp, newEnv := ltl.Match(oe.Child, tok)
	return OrEnvironment(oe.env, newOp), newEnv.Or(oe.env)
}

func (oe *orEnvironment) String() string {
	return fmt.Sprintf("OR_ENVIRONMENT(%s)", oe.env)
}

// Then is a temporal concatenation of its two arguments.  Its Match directs
// input Tokens to its left child until that Operator becomes nil, returning
// not Matching until that time, then directs input Tokens to its right child,
// returning the left child's final Environment ANDed with the right child's
// current Environment.
func Then(left, right ltl.Operator) ltl.Operator {
	if left == nil || right == nil {
		return nil
	}
	return &then{BinaryOperator{left, right}}
}

type then struct {
	BinaryOperator
}

func (t *then) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	op, env := ltl.Match(t.Left, tok)
	if op != nil {
		return Then(op, t.Right), env
	}
	return AndEnvironment(env, t.Right), ltl.NotMatching
}

func (t *then) String() string {
	return "THEN"
}

// Sequence is a temporal concatenation of all its arguments.  It is identical
// to a chain of Then operations, such that Sequence(a,b,c,..z) is equivalent
// to a THEN b THEN c THEN ... THEN z.
func Sequence(children ...ltl.Operator) ltl.Operator {
	return &sequence{
		NaryOperator{
			ChildSlice: children,
		},
	}
}

type sequence struct {
	NaryOperator
}

func (s *sequence) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	if len(s.ChildSlice) == 1 {
		return s.ChildSlice[0].Match(tok)
	}
	if len(s.ChildSlice) == 2 {
		return Then(s.ChildSlice[0], s.ChildSlice[1]).Match(tok)
	}
	return Then(s.ChildSlice[0], Sequence(s.ChildSlice[1:]...)).Match(tok)
}

func (s *sequence) String() string {
	return "SEQUENCE"
}

// Eventually is equivalent to its argument if that argument Matches at some
// point along its input Token stream.  Since its argument may need to accept
// multiple Tokens before resolving, Eventually may maintain an instance of
// its argument for each Token it accepts, returning the first to match.
// Because of this, Eventually can be expensive to use if not limited, such
// as with the Limit operation.
func Eventually(child ltl.Operator) ltl.Operator {
	if child == nil {
		return nil
	}
	return &eventually{UnaryOperator{child}}
}

type eventually struct {
	UnaryOperator
}

func (e *eventually) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	return StopAtFirstMatch(tok, Or(e.Child, Next(e)))
}

func (e *eventually) String() string {
	return "EVENTUALLY"
}

// Globally matches as long as its child matches.
func Globally(child ltl.Operator) ltl.Operator {
	return &globally{UnaryOperator{child}}
}

type globally struct {
	UnaryOperator
}

func (g *globally) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	op, env := g.Child.Match(tok)
	if op == nil {
		if !env.Matching() {
			return nil, env
		}
		return Globally(g.Child), env
	}
	return Or(op, Then(op, g)), env
}

func (g *globally) String() string {
	return "GLOBALLY"
}

// Until matches if its left argument holds until its right argument holds.   Its
// right argument must ultimately hold, but may hold immediately.  Once its right
// argument holds, Until terminates.
func Until(left, right ltl.Operator) ltl.Operator {
	if left == nil {
		return right
	}
	if right == nil {
		return nil
	}
	return &until{BinaryOperator{left, right}}
}

type until struct {
	BinaryOperator
}

func (u *until) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	return StopAtFirstMatch(tok, Or(u.Right, Then(u.Left, u)))
}

func (u *until) String() string {
	return "UNTIL"
}

// Release matches if its right child holds up to and including the time that
// its left child holds.  Its left child need never hold, in which case its
// right child must continually hold.
func Release(left, right ltl.Operator) ltl.Operator {
	return &release{BinaryOperator{left, right}}
}

type release struct {
	BinaryOperator
}

func (r *release) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	return Not(Until(Not(r.Left), Not(r.Right))).Match(tok)
}

func (r *release) String() string {
	return "RELEASE"
}
