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

package signals

import (
	"github.com/ilhamster/ltl/pkg/ltl"
	ops "github.com/ilhamster/ltl/pkg/operators"
	"strings"
	"testing"
)

func sm(s ...string) ltl.Operator {
	return NewMatcher(s...)
}

type testInput struct {
	input     string
	toks      []ltl.Token
	wantMatch bool
}

func parseToks(s string) []ltl.Token {
	var toks []ltl.Token
	for _, tok := range strings.Split(s, ";") {
		toks = append(toks, NewToken(strings.Split(tok, ",")...))
	}
	return toks
}

func m(s string) testInput {
	return testInput{s, parseToks(s), true}
}

func nm(s string) testInput {
	return testInput{s, parseToks(s), false}
}

func TestSignals(t *testing.T) {
	type testCase struct {
		op         ltl.Operator
		testInputs []testInput
	}
	tc := func(op ltl.Operator, testInputs ...testInput) testCase {
		return testCase{
			op:         op,
			testInputs: testInputs,
		}
	}
	tests := []testCase{
		tc(ops.Then(sm("a"), sm("b")),
			m("a,!b;!a,b"),
			nm("!a,b;!a,b"),
		),
		// Release and Until are dual.
		tc(ops.Release(sm("b"), sm("a")),
			m("a;a,b"),
			nm("a;!a,b"),
			m("a;a;a,b"),
			m("a;a;a"),
		),
		tc(ops.Not(ops.Until(ops.Not(sm("b")), ops.Not(sm("a")))),
			m("a;a,b"),
			nm("a;!a,b"),
			m("a;a;a,b"),
			m("a;a;a"),
		),
		tc(ops.Until(sm("a"), sm("b")),
			m("a;a;b"),
			nm("a;c"),
			nm("a;a;a"),
		),
		tc(ops.Not(ops.Release(ops.Not(sm("a")), ops.Not(sm("b")))),
			m("a;a;b"),
			nm("a;c"),
			nm("a;a;a"),
		),
		// Globally and Eventually are dual.
		tc(ops.Globally(sm("a")),
			m("a;a,b;a,!b"),
			nm("a;b"),
		),
		tc(ops.Not(ops.Eventually(ops.Not(sm("a")))),
			m("a;a,b;a,!b"),
			nm("a;b"),
		),
		tc(ops.Eventually(ops.Then(sm("a"), sm("b"))),
			m("c;d;b;a;b"),
			nm("c;d;b;a"),
		),
		tc(ops.Not(ops.Globally(ops.Then(ops.Not(sm("a")), ops.Not(sm("b"))))),
			m("c;d;b;a;b"),
			nm("c;d;b;a"),
		),
	}
	for _, test := range tests {
		for _, testInput := range test.testInputs {
			t.Run(ops.PrettyPrint(test.op, ops.Inline())+" <- "+testInput.input, func(t *testing.T) {
				op := test.op
				var env ltl.Environment
				for index, tok := range testInput.toks {
					if op == nil {
						t.Fatalf("op became nil")
					}
					op, env = ltl.Match(op, tok)
					if env.Err() != nil {
						t.Fatalf("at index %d unexpected error %s", index, env.Err())
					}
				}
				if testInput.wantMatch != env.Matching() {
					t.Fatalf("wanted match state %t, got %t", testInput.wantMatch, env.Matching())
				}
			})
		}
	}
}
