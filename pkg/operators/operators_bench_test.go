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
	rt "ltl/examples/runetoken"
	"ltl/pkg/ltl"
	"os"
	"runtime/pprof"
	"testing"
)

//                    0123456789 123456789 123456789 123456789 123456789 123456
const kStreamInput = "leg eg egg geleg lee gellel legg eglegleg gel egg egg leg"

const kNoProf = ""

func benchmarkEggLeg(b *testing.B, count int, profFile string) {
	if profFile != kNoProf {
		f, err := os.Create(profFile)
		if err != nil {
			b.Fatalf("Failed to open profile file: %s", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	tests := []struct {
		input          string
		op             ltl.Operator
		wantMatchCount int
	}{{
		kStreamInput,
		// The farthest-apart 'egg' and 'leg' are 21 characters apart.
		Eventually(Then(sm("egg"), Limit(21, Eventually(sm("leg"))))),
		6*count - 1, // First input has 5 matches; each subsequent has 6.
	}}
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			op := test.op
			var env ltl.Environment
			gotMatchCount := 0
			for n := 0; n < count*len(test.input); n++ {
				tok := rt.New(rune(test.input[n%len(test.input)]), n)
				op, env = op.Match(tok)
				if env.Matching() {
					gotMatchCount++
				}
				if op == nil {
					break
				}
			}
			if test.wantMatchCount != gotMatchCount {
				b.Fatalf("Expected %d matches, got %d", test.wantMatchCount, gotMatchCount)
			}
		}
	}
}

func BenchmarkEggLeg500(b *testing.B)   { benchmarkEggLeg(b, 500, kNoProf) }
func BenchmarkEggLeg5000(b *testing.B)  { benchmarkEggLeg(b, 5000, kNoProf) }
func BenchmarkEggLeg50000(b *testing.B) { benchmarkEggLeg(b, 50000, kNoProf) }
