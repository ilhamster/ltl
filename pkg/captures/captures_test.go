// Copyright 2021 Google LLC
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

package captures

import (
	"fmt"
	"testing"
)

type strTok string

func (st strTok) String() string {
	return string(st)
}

func (st strTok) EOI() bool {
	return false
}

type mapOp func(map[bool][]string)

func matching(strs ...string) mapOp {
	return func(m map[bool][]string) {
		m[true] = append(m[true], strs...)
	}
}

func notMatching(strs ...string) mapOp {
	return func(m map[bool][]string) {
		m[false] = append(m[false], strs...)
	}
}

func want(ops ...mapOp) map[bool][]string {
	m := map[bool][]string{}
	for _, op := range ops {
		op(m)
	}
	return m
}

func TestCaptures(t *testing.T) {
	for idx, test := range []struct {
		cap      *Captures
		captured map[bool][]string
	}{
		{nil, want()},
		{New().Capture(true, strTok("a")), want(matching("a"))},
		{New().Capture(false, strTok("a")), want(notMatching("a"))},
		{New().Capture(true, strTok("a")).Union(
			New().Capture(false, strTok("b")),
		), want(matching("a"), notMatching("b"))},
		{New().
			Capture(true, strTok("a")).
			Capture(false, strTok("b")).Not(),
			want(matching("b"), notMatching("a"))},
	} {
		t.Run(fmt.Sprintf("case %d", idx), func(t *testing.T) {
			for _, m := range []bool{true, false} {
				if caps := test.cap.Get(m); caps != nil {
					if len(test.captured[m]) != len(caps) {
						t.Fatalf("Got %d '%t' captures, expected %d", len(caps), m, len(test.captured[m]))
					}
					for _, cap := range test.captured[m] {
						if _, ok := caps[strTok(cap)]; !ok {
							t.Fatalf("Expected token %s to be captured for '%t', but it wasn't", cap, m)
						}
					}
				}
			}
		})
	}
}
