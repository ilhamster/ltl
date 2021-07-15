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

package integrationtest_test

import (
	"fmt"
	"github.com/ilhamster/ltl/examples/runetoken"
	"github.com/ilhamster/ltl/pkg/ltl"
	"os"
	"runtime/pprof"
	"testing"
)

// noProf should be used when no profiling is required.
const noProf = ""

const (
	input = "abacacbcdacadadbddcbabbdabcdadbabcbcaadbcab"
	// When input is repeated n times, matches 5*n+(n-1) times.
	streamExpr1 = "[$a<-] THEN (NOT [$a]) THEN [$a] THEN (NOT [$a]) THEN [$a]"
	// When input is repeated n times, matches 3*n+2*(n-1) times.
	streamExpr2 = "[$a<-] THEN [$b<-] THEN [$a] THEN [$b]"
	// When input is repeated n times, matches 2*n times.
	streamExpr3 = "[a] THEN [$a<-] THEN [b] THEN [$a]"
	// When input is repeated n times, matches 5n-1 times.
	onceExpr1 = "[$a<-] THEN (NOT [$a] AND [$b<-]) THEN EVENTUALLY ([$a] THEN [$b])"
)

// stream approximates matching against a continually streaming input by
// applying the provided input, repeated by the specified count, to the parsed
// provided expression.  A fresh instance of the expression is begun at each
// token, and is fed subsequent tokens until it becomes nil.  Matches are
// counted and, at the end, compared against an expected value.
// Maintaining multiple operators, from different starting points, is expensive.
func stream(b *testing.B, expr, input string, count int, wantMatch int, profFile string) {
	op, err := parse(expr)
	if err != nil {
		b.Fatalf("failed to parse expression: %s", err)
	}
	if len(profFile) > 0 {
		f, err := os.Create(fmt.Sprintf("%s.prof", profFile))
		if err != nil {
			b.Fatalf("Failed to open profile file: %s", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	for i := 0; i < b.N; i++ {
		var ops []ltl.Operator
		gotMatch := 0
		for n := 0; n < count*len(input); n++ {
			tok := runetoken.New(rune(input[n%len(input)]), n)
			ops = append(ops, op)
			var newOps []ltl.Operator
			for _, op := range ops {
				newOp, env := op.Match(tok)
				if ltl.IsErroring(env) {
					b.Fatalf("Unexpected error %s", env.Err())
				}
				if env.Matching() {
					gotMatch++
				}
				if newOp != nil {
					newOps = append(newOps, newOp)
				}
			}
			ops = newOps
		}
		if gotMatch != wantMatch {
			b.Fatalf("Expected %d matches, got %d", wantMatch, gotMatch)
		}
	}
}

// Once approximates matching against a continually streaming input by
// applying the provided input, repeated by the specified count, to the parsed
// provided expression.  A fresh instance of the expression is begun at each
// token, and is fed subsequent tokens until it becomes nil.  Matches are
// counted and, at the end, compared against an expected value.
// Maintaining multiple operators, from different starting points, is expensive.
func once(b *testing.B, expr, input string, count int, wantMatch int, profFile string) {
	op, err := parse(expr)
	if err != nil {
		b.Fatalf("failed to parse expression: %s", err)
	}
	if len(profFile) > 0 {
		f, err := os.Create(fmt.Sprintf("%s.prof", profFile))
		if err != nil {
			b.Fatalf("Failed to open profile file: %s", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	for i := 0; i < b.N; i++ {
		newOp := op
		gotMatch := 0
		for n := 0; n < count*len(input); n++ {
			tok := runetoken.New(rune(input[n%len(input)]), n)
			var env ltl.Environment
			newOp, env = newOp.Match(tok)
			if ltl.IsErroring(env) {
				b.Fatalf("Unexpected error %s", env.Err())
			}
			if env.Matching() {
				gotMatch++
			}
			if newOp == nil {
				break
			}
		}
		if gotMatch != wantMatch {
			b.Fatalf("Expected %d matches, got %d", wantMatch, gotMatch)
		}
	}
}

func BenchmarkStream_1_1_100(b *testing.B) {
	stream(b, streamExpr1, input, 100, 5*100+(100-1), noProf)
}

func BenchmarkStream_1_1_1000(b *testing.B) {
	stream(b, streamExpr1, input, 1000, 5*1000+(1000-1), noProf)
}

func BenchmarkStream_1_1_10000(b *testing.B) {
	stream(b, streamExpr1, input, 10000, 5*10000+(10000-1), noProf)
}

func BenchmarkStream_2_1_100(b *testing.B) {
	stream(b, streamExpr2, input, 100, 3*100+(2*(100-1)), noProf)
}

func BenchmarkStream_2_1_1000(b *testing.B) {
	stream(b, streamExpr2, input, 1000, 3*1000+(2*(1000-1)), noProf)
}

func BenchmarkStream_3_1_100(b *testing.B) {
	stream(b, streamExpr3, input, 100, 2*100, noProf)
}

func BenchmarkStream_3_1_1000(b *testing.B) {
	stream(b, streamExpr3, input, 1000, 2*1000, noProf)
}

func BenchmarkOnce_1_1_100(b *testing.B) {
	once(b, onceExpr1, input, 100, 5*100-1, noProf)
}

func BenchmarkOnce_1_1_1000(b *testing.B) {
	once(b, onceExpr1, input, 1000, 5*1000-1, noProf)
}

func BenchmarkOnce_1_1_10000(b *testing.B) {
	once(b, onceExpr1, input, 10000, 5*10000-1, noProf)
}
