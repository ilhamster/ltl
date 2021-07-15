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

package bindingenvironment

import (
	"fmt"
	"github.com/ilhamster/ltl/pkg/bindings"
	"testing"
)

func bs(bvs ...bindings.BoundValue) *bindings.Bindings {
	ret, err := bindings.New(bvs...)
	if err != nil {
		panic(fmt.Sprintf("Failed to create bindings: %s", err))
	}
	return ret
}

var result *bindings.Bindings
var (
	bA = bindings.String("a", "1")
	bB = bindings.Int("b", 2)

	bindAB = New(Bound(bs(bA, bB)))
	refA   = New(Referenced(bs(bA)))
	refB   = New(Referenced(bs(bB)))
)

func want(got, want *bindings.Bindings) {
	if !got.Eq(want) {
		panic(fmt.Sprintf("wanted %s, got %s", want, got))
	}
}

func Benchmark1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		env := bindAB.And(refA.And(refB))
		result = Bindings(env)
		want(result, bs(bA, bB))
	}
}

func Benchmark2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		env := bindAB.Or(refA.And(refB))
		result = Bindings(env)
		want(result, bs(bA, bB))
	}
}

func Benchmark3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		env := bindAB.And(bindAB).And(refA.And(refB))
		result = Bindings(env)
		want(result, bs(bA, bB))
	}
}
