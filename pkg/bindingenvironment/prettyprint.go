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
	"ltl/pkg/ltl"
	"sort"
	"strings"
)

// PrettyPrint pretty-prints bindingEnvironments for easier debugging.
// Non-binding Environments just get their matching state printed.
func PrettyPrint(env ltl.Environment, prefix ...string) {
	prefixStr := ""
	for _, p := range prefix {
		prefixStr = prefixStr + p
	}
	fmt.Print(prefixStr)
	if env == nil {
		fmt.Println("<nil>")
		return
	}
	switch v := env.(type) {
	case *binaryNode:
		t := ""
		switch v.t {
		case andNode:
			t = "AND"
		case orNode:
			t = "OR"
		}
		capStrs := []string{}
		caps := Captures(env).Get(env.Matching())
		if caps != nil {
			for cap := range caps {
				capStrs = append(capStrs, cap.String())
			}
			sort.Slice(capStrs, func(a, b int) bool {
				return capStrs[a] < capStrs[b]
			})
		}
		fmt.Printf("Binding %s (%t) (b: %s) (c: %s)\n", t, v.Matching(), v.bound, strings.Join(capStrs, ", "))
		PrettyPrint(v.left, prefixStr+"  ")
		PrettyPrint(v.right, prefixStr+"  ")
	case *BindingNode:
		fmt.Println(v.String())
	default:
		fmt.Println(ltl.State(env.Matching()))
	}
}
