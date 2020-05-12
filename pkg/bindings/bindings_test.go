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
    "testing"
)

func b(t *testing.T, bvs ...BoundValue) *Bindings {
    t.Helper()
    ret, err := New(bvs...)
    if err != nil {
        t.Fatalf("Failed to create Bindings: %s", err)
    }
    return ret
}

func bvl(bvs ...BoundValue) []BoundValue {
    return bvs
}

func TestBindingCreation(t *testing.T) {
    tests := []struct {
        bvs     []BoundValue
        want    *Bindings
        wantErr bool
    }{
        {bvl(String("a", "1")), b(t, String("a", "1")), false},
        {bvl(Int("a", 1)), b(t, Int("a", 1)), false},
        {bvl(String("a", "1"), String("a", "2")), nil, true},
        {bvl(String("a", "1"), String("a", "1")), nil, true},
        {bvl(Int("a", 1), Int("a", 2)), nil, true},
        {bvl(Int("a", 1), Int("a", 1)), nil, true},
        {bvl(String("a", "1"), Int("a", 1)), nil, true},
    }
    for _, test := range tests {
        t.Run(fmt.Sprintf("New(%v)", test.bvs), func(t *testing.T) {
            got, gotErr := New(test.bvs...)
            if gotErr == nil && test.wantErr {
                t.Fatalf("Expected an error but got none")
            }
            if gotErr != nil && !test.wantErr {
                t.Fatalf("Expected no error but got %s", gotErr)
            }
            if test.wantErr {
                return
            }
            if !got.Eq(test.want) {
                t.Fatalf("Wanted %s, got %s", test.want, got)
            }
        })
    }
}

func TestCombineBindings(t *testing.T) {
    tests := []struct {
        a, b, want *Bindings
        wantErr    bool
    }{
        // String-only bindings
        {b(t, String("a", "1")), b(t, String("b", "2")), b(t, String("a", "1"), String("b", "2")), false},
        {b(t, String("a", "1")), b(t, String("a", "1")), b(t, String("a", "1")), false},
        {b(t, String("a", "1")), b(t, String("a", "2")), nil, true},
        {b(t, String("a", "1"), String("b", "2")), b(t, String("c", "3")), b(t, String("a", "1"), String("b", "2"), String("c", "3")), false},
        // Int-only bindings
        {b(t, Int("a", 1)), b(t, Int("b", 2)), b(t, Int("a", 1), Int("b", 2)), false},
        {b(t, Int("a", 1)), b(t, Int("a", 1)), b(t, Int("a", 1)), false},
        {b(t, Int("a", 1)), b(t, Int("a", 2)), nil, true},
        {b(t, Int("a", 1), Int("b", 2)), b(t, Int("c", 3)), b(t, Int("a", 1), Int("b", 2), Int("c", 3)), false},
        // Mixed bindings
        {b(t, String("a", "1")), b(t, Int("b", 2)), b(t, String("a", "1"), Int("b", 2)), false},
        {b(t, Int("a", 1)), b(t, String("a", "1")), nil, true},
        {b(t, Int("a", 1), String("b", "2")), b(t, Int("c", 3), String("d", "4")), b(t, Int("a", 1), String("b", "2"), Int("c", 3), String("d", "4")), false},
    }
    for _, test := range tests {
        t.Run(fmt.Sprintf("Combine(%s, %s)", test.a, test.b), func(t *testing.T) {
            a, gotErr := test.a.Combine(test.b)
            if gotErr == nil && test.wantErr {
                t.Fatalf("Expected an error but got none")
            }
            if gotErr != nil && !test.wantErr {
                t.Fatalf("Expected no error but got %s", gotErr)
            }
            if test.wantErr {
                return
            }
            if !a.Eq(test.want) {
                t.Fatalf("Wanted %s, got %s", test.want, a)
            }
        })
    }
}

func TestSatisfyBindings(t *testing.T) {
    tests := []struct {
        a, b, want    *Bindings
        wantSatisfied bool
    }{
        // String-only bindings
        {b(t, String("a", "1")), b(t, String("b", "2")), b(t, String("a", "1")), true},
        {b(t, String("a", "1")), b(t, String("a", "1")), nil, true},
        {b(t, String("a", "1")), b(t, String("a", "2")), nil, false},
        {b(t, String("a", "1"), String("b", "2")), b(t, String("a", "1")), b(t, String("b", "2")), true},
        // Int-only bindings
        {b(t, Int("a", 1)), b(t, Int("b", 2)), b(t, Int("a", 1)), true},
        {b(t, Int("a", 1)), b(t, Int("a", 1)), nil, true},
        {b(t, Int("a", 1)), b(t, Int("a", 2)), nil, false},
        {b(t, Int("a", 1), Int("b", 2)), b(t, Int("a", 1)), b(t, Int("b", 2)), true},
        // Mixed bindings
        {b(t, Int("a", 1), String("b", "2")), b(t, Int("c", 3)), b(t, Int("a", 1), String("b", "2")), true},
        {b(t, Int("a", 1), String("b", "2")), b(t, Int("a", 1), String("b", "2")), nil, true},
        {b(t, Int("a", 1)), b(t, String("a", "1")), nil, false},
    }
    for _, test := range tests {
        t.Run(fmt.Sprintf("Satisfy(%s, %s)", test.a, test.b), func(t *testing.T) {
            a, gotSatisfied := test.a.Satisfy(test.b)
            if gotSatisfied != test.wantSatisfied {
                t.Fatalf("= %t, wanted %t", gotSatisfied, test.wantSatisfied)
            }
            if test.wantSatisfied && !a.Eq(test.want) {
                t.Fatalf("Wanted %s, got %s", test.want, a)
            }
        })
    }
}

func TestCompareBindings(t *testing.T) {
    tests := []struct {
        a, b *Bindings
        want bool
    }{
        // String-only bindings
        {nil, b(t, String("a", "1")), false},
        {b(t, String("a", "1")), nil, false},
        {nil, nil, true},
        {b(t, String("a", "1")), b(t, String("b", "2")), false},
        {b(t, String("b", "2")), b(t, String("a", "1")), false},
        {b(t, String("a", "1")), b(t, String("a", "1")), true},
        {b(t, String("a", "1")), b(t, String("a", "1"), String("b", "2")), false},
        {b(t, String("a", "1"), String("b", "2")), b(t, String("a", "1")), false},
        {b(t, String("a", "1"), String("b", "2")), b(t, String("b", "2"), String("a", "1")), true},
        // Int-only bindings.
        {nil, b(t, Int("a", 1)), false},
        {b(t, Int("a", 1)), nil, false},
        {nil, nil, true},
        {b(t, Int("a", 1)), b(t, Int("b", 2)), false},
        {b(t, Int("b", 2)), b(t, Int("a", 1)), false},
        {b(t, Int("a", 1)), b(t, Int("a", 1)), true},
        {b(t, Int("a", 1)), b(t, Int("a", 1), Int("b", 2)), false},
        {b(t, Int("a", 1), Int("b", 2)), b(t, Int("a", 1)), false},
        {b(t, Int("a", 1), Int("b", 2)), b(t, Int("b", 2), Int("a", 1)), true},
        // Mixed bindings.
        {b(t, Int("a", 1)), b(t, String("a", "1")), false},
        {b(t, Int("a", 1), String("b", "2")), b(t, String("b", "2"), Int("a", 1)), true},
    }
    for _, test := range tests {
        t.Run(fmt.Sprintf("Compare(%s, %s)", test.a, test.b), func(t *testing.T) {
            got := test.a.Eq(test.b)
            if got != test.want {
                t.Fatalf("= %t, wanted %t", got, test.want)
            }
        })
    }
}
