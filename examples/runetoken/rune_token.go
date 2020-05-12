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

// Package runetoken provides an ltl.Token containing a rune and a unique index.
package runetoken

import "fmt"

// RuneToken implements ltl.Token for rune tokens with indices.
type RuneToken struct {
	r     rune
	index int
}

// Returns a new RuneToken with the provided rune and index.
func New(r rune, index int) *RuneToken {
	return &RuneToken{r, index}
}

func (st *RuneToken) EOI() bool {
	return false
}

func (st *RuneToken) Value() rune {
	return st.r
}

func (st *RuneToken) Index() int {
	return st.index
}

func (st *RuneToken) String() string {
	return fmt.Sprintf("%s (%d)", string(st.r), st.index)
}
