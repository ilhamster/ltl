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

package test

import (
	"bufio"
	"fmt"
	"io"
	rt "ltl/examples/runetoken"
	smatch "ltl/examples/stringmatcher"
	be "ltl/pkg/bindingenvironment"
	"ltl/pkg/bindings"
	"ltl/pkg/ltl"
	ops "ltl/pkg/operators"
	"ltl/pkg/parser"
	"strings"
	"testing"
)

func parse(s string) (ltl.Operator, error) {
	l, err := parser.NewLexer(parser.DefaultTokens,
		smatch.Generator(smatch.Capture(true)),
		bufio.NewReader(strings.NewReader(s)))
	if err != nil {
		return nil, err
	}
	return parser.ParseLTL(l)
}

type testInput struct {
	input        string
	wantBindings *bindings.Bindings
	wantMatch    bool
	wantErr      bool
	wantIndices  map[int]struct{}
}

func b(args ...string) func(*testInput) {
	if len(args)%2 == 1 {
		panic("b() requires an even-length argument")
	}
	bvs := make([]bindings.BoundValue, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		bvs = append(bvs, bindings.String(args[i], args[i+1]))
	}
	ret, err := bindings.New(bvs...)
	if err != nil {
		panic("failed to create bindings: " + err.Error())
	}
	return func(ti *testInput) {
		ti.wantBindings = ret
	}
}

func i(args ...int) func(*testInput) {
	return func(ti *testInput) {
		ti.wantIndices = map[int]struct{}{}
		for _, idx := range args {
			ti.wantIndices[idx] = struct{}{}
		}
	}
}

const (
	notMatching bool = false
	matching    bool = true
)

func m(input string, opts ...func(*testInput)) *testInput {
	ti := &testInput{
		input:     input,
		wantMatch: matching,
	}
	for _, opt := range opts {
		opt(ti)
	}
	return ti
}

func nm(input string) *testInput {
	return &testInput{
		input:        input,
		wantMatch:    notMatching,
		wantBindings: nil,
	}
}

func err(input string) *testInput {
	return &testInput{
		input:     input,
		wantMatch: notMatching,
		wantErr:   true,
	}
}

func expect(op ltl.Operator, input *testInput, t *testing.T) {
	t.Helper()
	var env ltl.Environment
	reader := strings.NewReader(input.input)
	for index := 0; ; index++ {
		r, _, err := reader.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Read error: %s", err)
		}
		tok := rt.New(r, index)
		if op == nil {
			env = ltl.NotMatching
			break
		}
		op, env = ltl.Match(op, tok)
	}
	if env.Err() != nil && !input.wantErr {
		t.Fatalf("unexpected error %s", env.Err())
	}
	if env.Err() == nil && input.wantErr {
		t.Fatalf("wanted error %s but got none", env.Err())
	}
	gotMatch := env.Matching()
	if input.wantMatch != gotMatch {
		// be.PrettyPrint(env)
		t.Fatalf("Wanted match state %t, got %t", input.wantMatch, env.Matching())
	}
	if gotMatch == notMatching {
		return
	}
	if !input.wantBindings.Eq(be.Bindings(env)) {
		t.Fatalf("Wanted Environment bindings %s, got %s", input.wantBindings, be.Bindings(env))
	}
	captured := be.Captures(env)
	if len(input.wantIndices) != len(captured) {
		fmt.Println(op)
		be.PrettyPrint(env)
		t.Fatalf("Wanted %d captures, got %d", len(input.wantIndices), len(captured))
	}
	if input.wantIndices != nil && captured != nil {
		for tok := range captured {
			if rt, ok := tok.(*rt.RuneToken); ok {
				if _, ok := input.wantIndices[rt.Index()]; !ok {
					t.Fatalf("Captured unexpected index %d", rt.Index())
				}
			}
		}

	}
	if env.Err() != nil {
		t.Fatalf("MatchOnBindings yielded unexpected error %s", env.Err())
	}
	if input.wantMatch != env.Matching() {
		t.Errorf("wanted match state %t, got %t", input.wantMatch, env.Matching())
		return
	}
}

// Tests a variety of interesting expressions.
func TestExpressions(t *testing.T) {
	type testCase struct {
		opStr     string
		inputSets []*testInput
		op        ltl.Operator
	}
	tc := func(opStr string, inputSets ...*testInput) testCase {
		op, err := parse(opStr)
		if err != nil {
			t.Fatalf("Failed to parse: %s", err)
		}
		return testCase{opStr, inputSets, op}
	}
	tests := []testCase{
		tc("[1]",
			m("1", i(0)),
			nm("2"),
		),
		tc("[1] THEN [2]",
			m("12", i(0, 1)),
			nm("21"),
			nm("11"),
		),
		tc("[1] THEN [2] THEN EVENTUALLY [3]",
			m("12443", i(0, 1, 4)),
			m("12223", i(0, 1, 4)),
			nm("13"),
			nm("3"),
			nm("12"),
		),
		tc("([a] OR [b]) UNTIL NOT ([b] OR [a])",
			m("ababc", i(0, 1, 2, 3, 4)),
			nm("bb"),
			nm("abab"),
			m("c", i(0)),
		),
		tc("[$a<-] THEN [2] THEN [$a]",
			m("121", b("a", "1"), i(0, 1, 2)),
			nm("321"),
			nm("12"),
		),
		tc("[$a<-] THEN (([1] THEN [2]) UNTIL [$a])",
			m("312123", b("a", "3"), i(0, 1, 2, 3, 4, 5)),
			m("1121", b("a", "1"), i(0, 1, 2, 3)),
			m("11", b("a", "1"), i(0, 1)),
			nm("1122"),
		),
		tc("[$a<-] THEN NOT [$a]",
			m("12", b("a", "1"), i(0, 1)),
			nm("11"),
		),
		tc("[$a<-] THEN EVENTUALLY NOT [$a]",
			m("12", b("a", "1"), i(0, 1)),
			m("112", b("a", "1"), i(0, 2)),
			nm("111"),
		),
		tc("([$a<-] AND [b]) THEN (([e] UNTIL [f]) THEN [$a])",
			m("beefb", b("a", "b"), i(0, 1, 2, 3, 4)),
			nm("beefa"),
		),
		tc("[abc] THEN [def]",
			m("abcdef", i(2, 5)),
			nm("nope"),
		),
		tc("[$a<-] THEN ([$b<-] AND NOT [$a]) THEN [$a] THEN [$b] THEN [$a]",
			m("12121", b("a", "1", "b", "2"), i(0, 1, 2, 3, 4)),
			nm("11111"),
			nm("12111"),
		),
		tc("(EVENTUALLY [1]) UNTIL [2]",
			m("1313312", i(0, 2, 5, 6)),
			nm("131331)"),
			m("2", i(0)),
		),
		tc("(EVENTUALLY [1]) AND (EVENTUALLY [2]) AND EVENTUALLY [3]",
			m("414342", i(1, 3, 5)),
			nm("41434"),
			nm("444"),
		),
		tc("(EVENTUALLY [$a<-]) AND EVENTUALLY ([$a] THEN [$a])",
			m("111", b("a", "1"), i(0, 1, 2)),
			nm("13"),
		),
		// These are risky queries, as they can produce multiple bound values
		// for binding keys.  Not recommended.

		// This query relies on short-circuiting to avoid multiply-binding $a.
		// Nonetheless, it's not good practice.
		tc("[$a<-] UNTIL NOT [$a]",
			nm("bbbb"),
			m("bbba", b("a", "b"), i(0, 1, 2, 3)),
			m("12", b("a", "1"), i(0, 1)),
			nm("11"),
			// $a is unbound, so no match.
			nm("1"),
		),

		tc("[$a<-] THEN ([$b<-] UNTIL [$a])",
			nm("abb"),
			err("abca"),
		),
		tc("[$a<-] THEN [$b<-] THEN ([$b] UNTIL [$a])",
			nm("abb"),
			nm("abca"),
			m("abba", b("a", "a", "b", "b"), i(0, 1, 2, 3)),
		),
		tc("[$a<-] THEN ([$b<-] UNTIL [$a])",
			m("abba", b("a", "a", "b", "b"), i(0, 1, 2, 3)),
			m("ccc", b("a", "c", "b", "c"), i(0, 1, 2)),
			err("cabc"),
		),
		tc("[$a<-] THEN [$a<-]",
			m("11", b("a", "1"), i(0, 1)),
			err("12"),
		),
	}
	for _, test := range tests {
		for _, inputSet := range test.inputSets {
			t.Run(fmt.Sprintf("%s <- %s", ops.PrettyPrint(test.op, ops.Inline()), inputSet.input), func(t *testing.T) {
				expect(test.op, inputSet, t)
			})
		}
	}
}

// Tests that different formulae that should be equivalent actually are.
func TestEquivalentFormulae(t *testing.T) {
	type testCase struct {
		description string
		opStrs      []string
		inputSets   []*testInput
		ops         []ltl.Operator
	}
	tc := func(description string, opStrs []string, inputSets ...*testInput) testCase {
		ops := []ltl.Operator{}
		for _, opStr := range opStrs {
			op, err := parse(opStr)
			if err != nil {
				t.Fatalf("Failed to parse: %s", err)
			}
			ops = append(ops, op)
		}
		return testCase{description, opStrs, inputSets, ops}
	}
	tests := []testCase{
		tc("DeMorgan's Law: NOT OR, literals",
			[]string{
				"[1] OR [2]",
				"NOT (NOT [1] AND NOT [2])",
			},
			m("1", i(0)),
			m("2", i(0)),
			nm("3"),
		),
		tc("DeMorgan's Law: NOT OR, bindings",
			[]string{
				"[$a<-] OR [2]",
				"NOT ((NOT [$a<-]) AND NOT [2])",
			},
			m("1", b("a", "1"), i(0)),
			m("2", b("a", "2"), i(0)),
			m("3", b("a", "3"), i(0)),
		),
		tc("DeMorgan's Law: NOT AND, literals",
			[]string{
				"([1] THEN [.]) AND ([.] THEN [2])",
				"NOT (NOT ([1] THEN [.]) OR NOT ([.] THEN [2]))",
			},
			m("12", i(0, 1)),
			nm("22"),
			nm("1"),
		),
		tc("DeMorgan's Law: NOT AND, bindings",
			[]string{
				"([$a<-] THEN [.]) AND ([.] THEN [2])",
				"NOT (NOT ([$a<-] THEN [.]) OR NOT ([.] THEN [2]))",
				// "[$a<-] AND NEXT [2]",
				// "NOT (NOT [$a<-] OR NOT NEXT [2])",
			},
			m("12", b("a", "1"), i(0, 1)),
			m("22", b("a", "2"), i(0, 1)),
			nm("21"),
			nm("1"),
		),
		tc("UNTIL-RELEASE duality",
			[]string{
				"([a] OR [b]) UNTIL ([b] OR [c])",
				"NOT ((NOT ([a] OR [b])) RELEASE NOT ([b] OR [c]))",
			},
			m("ab", i(0, 1)),
			m("aaab", i(0, 1, 2, 3)),
			nm("ca"),
		),
		tc("RELEASE-UNTIL duality",
			[]string{
				"([b] OR [c]) RELEASE ([a] OR [b])",
				"NOT ((NOT ([b] OR [c])) UNTIL NOT ([a] OR [b]))",
			},
			m("abc", i(1)),
			m("bc", i(0)),
			nm("cb"),
		),
		tc("THEN-OR distributivity, literals",
			[]string{
				"[a] THEN ([b] OR [c])",
				"([a] THEN [b]) OR ([a] THEN [c])",
			},
			m("ab", i(0, 1)),
			m("ac", i(0, 1)),
			nm("b"),
		),
		tc("THEN-OR distributivity, bindings",
			[]string{
				"[$a<-] THEN ([2] OR [$a])",
				"([$a<-] THEN [2]) OR ([$a<-] THEN [$a])",
			},
			m("12", b("a", "1"), i(0, 1)),
			nm("13"),
			m("11", b("a", "1"), i(0, 1)),
		),
		tc("THEN-AND distributivity",
			[]string{
				"[$a<-] THEN ([$a] AND [b])",
				"([$a<-] THEN [$a]) AND ([$a<-] THEN [b])",
			},
			m("bb", b("a", "b"), i(0, 1)),
			nm("ab"),
		),
		tc("EVENTUALLY-OR distributivity, literals",
			[]string{
				"EVENTUALLY ([a] OR [b])",
				"EVENTUALLY [a] OR EVENTUALLY [b]",
			},
			m("cca", i(2)),
			m("ccb", i(2)),
			nm("ccc"),
		),
		tc("EVENTUALLY-OR distributivity, bindings",
			[]string{
				"[$a<-] THEN EVENTUALLY ([$a] OR [2])",
				"[$a<-] THEN ((EVENTUALLY [$a]) OR (EVENTUALLY [2]))",
			},
			m("131", b("a", "1"), i(0, 2)),
			nm("133"),
			m("132", b("a", "1"), i(0, 2)),
		),
		tc("UNTIL-OR distributivity, literals",
			[]string{
				"[a] UNTIL ([b] OR [c])",
				"([a] UNTIL [b]) OR ([a] UNTIL [c])",
			},
			m("aab", i(0, 1, 2)),
			m("aac", i(0, 1, 2)),
			nm("aaa"),
		),
		tc("UNTIL-OR distributivity, bindings",
			[]string{
				"[$a<-] THEN ([1] UNTIL ([2] OR [$a]))",
				"[$a<-] THEN (([1] UNTIL [2]) OR ([1] UNTIL [$a]))",
			},
			m("312", b("a", "3"), i(0, 1, 2)),
			nm("3"),
			m("313", b("a", "3"), i(0, 1, 2)),
		),
		tc("UNTIL-AND distributivity",
			[]string{
				"(NOT [a] AND NOT [b]) UNTIL [a]",
				"(NOT [a] UNTIL [a]) AND (NOT [b] UNTIL [a])",
			},
			m("cca", i(0, 1, 2)),
			m("a", i(0)),
			nm("ccc"),
		),
	}
	for _, test := range tests {
		for _, inputSet := range test.inputSets {
			for idx, op := range test.ops {
				t.Run(fmt.Sprintf("%s: %s <- %s", test.description, test.opStrs[idx], inputSet.input), func(t *testing.T) {
					expect(op, inputSet, t)
				})
			}
		}
	}
}
