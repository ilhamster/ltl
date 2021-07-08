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
	"ltl/pkg/bindings"
	"ltl/pkg/ltl"
	"testing"
)

func sb(args ...string) *bindings.Bindings {
	if len(args)%2 == 1 {
		panic("nsb requires an even number of string arguments")
	}
	bvs := make([]bindings.BoundValue, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		bvs = append(bvs, bindings.String(args[i], args[i+1]))
	}
	ret, err := bindings.New(bvs...)
	if err != nil {
		panic("failed to create bindings: " + err.Error())
	}
	return ret
}

func bind(args ...string) bindingEnvironment {
	return New(Matching(true), Bound(sb(args...)))
}

func ref(args ...string) bindingEnvironment {
	return New(Matching(true), Referenced(sb(args...)))
}

func rwb(refs, bindings *bindings.Bindings) bindingEnvironment {
	return New(Matching(true), Bound(bindings), Referenced(refs))
}

func TestApplyBindings(t *testing.T) {
	tests := []struct {
		n       bindingEnvironment
		b       *bindings.Bindings
		want    bindingEnvironment
		wantErr bool
	}{
		{ref("a", "1"), sb("a", "1"), bind("a", "1"), false},
		{bind("a", "1"), sb("b", "2"), bind("a", "1", "b", "2"), false},
		{bind("a", "1"), sb("a", "2"), nil, true},
		{rwb(sb("a", "1"), sb("b", "2")), sb("a", "1"), bind("a", "1", "b", "2"), false},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s <- %s", test.n, test.b), func(t *testing.T) {
			got := test.n.applyBindings(test.b)
			if test.wantErr && got.Err() == nil {
				t.Fatalf("Wanted error but got none")
			}
			if !test.wantErr && got.Err() != nil {
				t.Fatalf("Wanted no error but got %s", got.Err())
			}
			if test.wantErr {
				return
			}
			if _, ok := merge(test.want, got); !ok {
				t.Fatalf("Wanted %s, got %s", test.want, got)
			}
		})
	}
}

func TestBindingCombinations(t *testing.T) {
	tests := []struct {
		env       ltl.Environment
		wantMatch bool
		wantErr   bool
	}{
		{bind("a", "1"), true, false},
		{bind("a", "1").And(bind("b", "2")), true, false},
		{bind("a", "1").And(bind("a", "2")), false, true},
		{bind("a", "1").And(ref("a", "1")), true, false},
		{bind("a", "1").And(ref("a", "2")), false, false},
		{bind("a", "1").And(bind("b", "2")).And(ref("a", "1")), true, false},
		{bind("a", "1").And(bind("b", "2")).And(ref("a", "1")).And(ref("b", "2")), true, false},
		{bind("a", "1").And(ref("a", "1")).And(ref("b", "2")), false, false},
		{bind("a", "1").And(bind("b", "2")).And((ref("a", "1")).Or(ref("b", "2"))), true, false},
		{bind("a", "1").And(bind("b", "2")).And(bind("b", "2")).And(ref("a", "1")), true, false},
	}
	for idx, test := range tests {
		t.Run(fmt.Sprintf("test case %d", idx), func(t *testing.T) {
			if test.env.Matching() != test.wantMatch {
				t.Fatalf("Wanted %s.Matching() to be %t, got %t", test.env, test.wantMatch, test.env.Matching())
			}
			if test.env.Err() == nil && test.wantErr {
				t.Fatalf("Wanted %s to have error but got none", test.env)
			}
			if test.env.Err() != nil && !test.wantErr {
				t.Fatalf("Wanted %s to have no error but got %s", test.env, test.env.Err())
			}
		})
	}
}

type strTok string

func (st strTok) String() string {
	return string(st)
}

func (st strTok) EOI() bool {
	return false
}

func cap(matching bool, args ...string) bindingEnvironment {
	toks := []ltl.Token{}
	for _, arg := range args {
		toks = append(toks, strTok(arg))
	}
	return New(Matching(matching), Captured(toks...))
}

func strs(strs ...string) map[string]struct{} {
	if len(strs) == 0 {
		return nil
	}
	ret := map[string]struct{}{}
	for _, str := range strs {
		ret[str] = struct{}{}
	}
	return ret
}

func TestCaptures(t *testing.T) {
	tests := []struct {
		env          ltl.Environment
		wantCaptures map[string]struct{}
	}{
		{cap(true, "a").Or(cap(true, "b")), strs("a", "b")},
		{cap(false, "a").Or(cap(true, "b")), strs("b")},
	}
	for idx, test := range tests {
		t.Run(fmt.Sprintf("test case %d", idx), func(t *testing.T) {
			PrettyPrint(test.env)
			gotCaptures := Captures(test.env)
			if len(gotCaptures) != len(test.wantCaptures) {
				t.Fatalf("Wanted %d captures, got %d", len(test.wantCaptures), len(gotCaptures))
			}
			if gotCaptures != nil && test.wantCaptures != nil {
				for cap := range gotCaptures {
					if st, ok := cap.(strTok); !ok {
						t.Fatalf("Expected all captures to be strToks")
					} else {
						if _, ok := test.wantCaptures[st.String()]; !ok {
							t.Fatalf("Unexpected capture %s", st)
						}
					}
				}
			}
		})
	}
}
