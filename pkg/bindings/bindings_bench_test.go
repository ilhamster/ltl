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

package bindings

import (
	"fmt"
	"os"
	"runtime/pprof"
	"testing"
)

// noProf should be used when no profiling is required.
const noProf = ""

func createBindings(b *testing.B, boundValues [][]BoundValue) []*Bindings {
	bindings := make([]*Bindings, 0, len(boundValues))
	for _, bvs := range boundValues {
		binding, err := New(bvs...)
		if err != nil {
			b.Fatalf("Failed to create Bindings: %s", err)
		}
		bindings = append(bindings, binding)
	}
	return bindings
}

type testType int

const (
	combine testType = iota
	satisfy
)

func (tt testType) String() string {
	switch tt {
	case combine:
		return "combine"
	case satisfy:
		return "satisfy"
	default:
		return "unknown"
	}
}

// The iteration cursor is package-global to prevent bench being optimized out.
var cursor *Bindings

func bench(b *testing.B, tt testType, boundValues [][]BoundValue, profFile string) {
	bindings := createBindings(b, boundValues)
	if len(bindings) <= 1 {
		b.Fatalf("At least 2 bindings must be provided.")
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
		cursor = bindings[0]
		for _, binding := range bindings[1:] {
			var newCursor *Bindings
			var err error
			switch tt {
			case combine:
				newCursor, err = cursor.Combine(binding)
			case satisfy:
				newCursor, _ = cursor.Satisfy(binding)
			default:
				b.Fatalf("Unsupported test type %v", tt)
			}
			if err != nil {
				b.Fatalf("Failed to %s %s and %s: %s", tt, cursor, binding, err)
			}
			cursor = newCursor
		}
	}
}

var shortKeyInts = [][]BoundValue{
	{
		Int("a", 1),
		Int("b", 2),
	}, {
		Int("c", 3),
		Int("d", 4),
	}, {
		Int("e", 5),
		Int("a", 1),
	},
}

var longKeyInts = [][]BoundValue{
	{
		Int("Phenomenal", 1),
		Int("Exornitant", 2),
	}, {
		Int("Remarkable", 3),
		Int("Excessive", 4),
	}, {
		Int("Preposterous", 5),
		Int("Phenomenal", 1),
	},
}

var longKeyShortStrings = [][]BoundValue{
	{
		String("Phenomenal", "a"),
		String("Exornitant", "b"),
	}, {
		String("Remarkable", "c"),
		String("Excessive", "d"),
	}, {
		String("Preposterous", "e"),
		String("Phenomenal", "a"),
	},
}

var shortKeyShortStrings = [][]BoundValue{
	{
		String("a", "a"),
		String("b", "b"),
	}, {
		String("c", "c"),
		String("d", "d"),
	}, {
		String("e", "e"),
		String("a", "a"),
	},
}

var longKeyLongStrings = [][]BoundValue{
	{
		String("Phenomenal", "ridiculous"),
		String("Exornitant", "overwhelming"),
	}, {
		String("Remarkable", "beyond the pale"),
		String("Excessive", "genuinely absurd"),
	}, {
		String("Preposterous", "intolerable"),
		String("Phenomenal", "ridiculous"),
	},
}

var shortKeyLongStrings = [][]BoundValue{
	{
		String("a", "ridiculous"),
		String("b", "overwhelming"),
	}, {
		String("c", "beyond the pale"),
		String("d", "genuinely absurd"),
	}, {
		String("e", "intolerable"),
		String("a", "ridiculous"),
	},
}

func BenchmarkCombineShortKeyInts(b *testing.B) {
	bench(b, combine, shortKeyInts, noProf)
}

func BenchmarkSatisfyShortKeyInts(b *testing.B) {
	bench(b, satisfy, shortKeyInts, noProf)
}

func BenchmarkCombineLongKeyInts(b *testing.B) {
	bench(b, combine, longKeyInts, noProf)
}

func BenchmarkSatisfyLongKeyInts(b *testing.B) {
	bench(b, satisfy, longKeyInts, noProf)
}

func BenchmarkCombineShortKeyShortStrings(b *testing.B) {
	bench(b, combine, shortKeyShortStrings, noProf)
}

func BenchmarkSatisfyShortKeyShortStrings(b *testing.B) {
	bench(b, satisfy, shortKeyShortStrings, noProf)
}

func BenchmarkCombineLongKeyShortStrings(b *testing.B) {
	bench(b, combine, longKeyShortStrings, noProf)
}

func BenchmarkSatisfyLongKeyShortStrings(b *testing.B) {
	bench(b, satisfy, longKeyShortStrings, noProf)
}

func BenchmarkCombineShortKeyLongStrings(b *testing.B) {
	bench(b, combine, shortKeyLongStrings, noProf)
}

func BenchmarkSatisfyShortKeyLongStrings(b *testing.B) {
	bench(b, satisfy, shortKeyLongStrings, noProf)
}

func BenchmarkCombineLongKeyLongStrings(b *testing.B) {
	bench(b, combine, longKeyLongStrings, noProf)
}

func BenchmarkSatisfyLongKeyLongStrings(b *testing.B) {
	bench(b, satisfy, longKeyLongStrings, noProf)
}
