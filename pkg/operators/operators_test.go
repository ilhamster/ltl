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
	rtok "github.com/ilhamster/ltl/examples/runetoken"
	smatch "github.com/ilhamster/ltl/examples/stringmatcher"
	"github.com/ilhamster/ltl/pkg/ltl"
	"testing"
)

var capture = true

func sm(s string) ltl.Operator {
	return smatch.New(s, smatch.Capture(capture))
}

type testInput struct {
	input     string
	wantMatch bool
}

func m(s string) testInput {
	return testInput{s, true}
}

func nm(s string) testInput {
	return testInput{s, false}
}

func TestOperators(t *testing.T) {
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
		tc(sm("a"),
			m("a"), nm("b")),
		tc(Not(sm("b")),
			m("a"), nm("b")),
		tc(Or(sm("a"), sm("b")),
			m("a"), m("b"), nm("c")),
		tc(Then(sm("a"), sm("b")),
			m("ab"), nm("aa"), nm("a"), nm("c")),
		tc(Then(sm("a"), Not(sm("b"))),
			m("aa")),
		tc(Eventually(sm("b")),
			m("aaab")),
		tc(Eventually(Not(sm("b"))),
			m("bbba")),
		tc(Then(sm("a"), Eventually(sm("b"))),
			m("aaab")),
		tc(Then(sm("a"), Then(sm("b"), Eventually(sm("c")))),
			m("abddc")),
		tc(Then(Eventually(sm("a")), Eventually(sm("b"))),
			m("caacb")),
		tc(Eventually(Then(sm("a"), Eventually(sm("b")))),
			m("caacb")),
		tc(Eventually(Then(sm("a"), sm("b"))),
			nm("ba"), m("caab")),
		tc(Eventually(Or(sm("a"), sm("b"))),
			m("cccb"),
			m("ccca"),
			nm("ccc")),
		tc(Until(Or(sm("a"), sm("b")), Not(sm("b"))),
			m("bbba"), nm("bb")),
		tc(Until(sm("a"), Or(sm("b"), sm("c"))),
			m("aab")),
		tc(Until(sm("a"), Then(sm("b"), sm("c"))),
			m("abc"), m("aabc"), nm("aac")),
		tc(Until(Then(sm("a"), sm("b")), sm("c")),
			m("abc"), m("ababc")),
		tc(Then(Sequence(sm("e"), sm("g"), sm("g")), Eventually(Sequence(sm("l"), sm("e"), sm("g")))),
			m("egg leg"), nm("egg"), nm("egg le")),
		tc(Limit(5, Then(sm("a"), Eventually(sm("b")))),
			m("ab"), m("aaaab"), nm("aaaaa")),
	}
	for _, test := range tests {
		for _, testInput := range test.testInputs {
			t.Run(PrettyPrint(test.op, Inline())+" <- "+testInput.input, func(t *testing.T) {
				op := test.op
				var env ltl.Environment
				var toks []ltl.Token
				for idx, ch := range testInput.input {
					// for idx, ch := range strings.Split(testInput.input, "") {
					toks = append(toks, rtok.New(ch, idx))
				}
				for index, tok := range toks {
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
