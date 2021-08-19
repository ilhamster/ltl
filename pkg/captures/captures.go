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

// Package captures provides a utility type for capturing tokens participating
// in matches.
package captures

import "github.com/ilhamster/ltl/pkg/ltl"

// Captures stores sets of tokens captured by Environments.
type Captures struct {
	// Caps stores two sets of captured tokens: one captured if the Environment
	// matches, and one captured if it does not match.
	caps map[bool]map[ltl.Token]struct{}
}

// New returns a new, empty Captures set.
func New() *Captures {
	return &Captures{
		caps: map[bool]map[ltl.Token]struct{}{
			true:  nil,
			false: nil,
		},
	}
}

// Get returns the set of tokens captured under the provided matching state.
// The returned map may be nil.
func (c *Captures) Get(matching bool) map[ltl.Token]struct{} {
	if c == nil {
		return nil
	}
	return c.caps[matching]
}

// Capture captures the provided set of tokens under the specified matching
// state.  It returns itself, for chaining.
func (c *Captures) Capture(matching bool, toks ...ltl.Token) *Captures {
	if c.caps[matching] == nil {
		c.caps[matching] = map[ltl.Token]struct{}{}
	}
	for _, tok := range toks {
		c.caps[matching][tok] = struct{}{}
	}
	return c
}

// Union returns a new Capture comprised of the union of the receiver and the
// argument.
func (c *Captures) Union(oc *Captures) *Captures {
	if c == nil {
		return oc
	}
	if oc == nil {
		return c
	}
	ret := &Captures{map[bool]map[ltl.Token]struct{}{}}

	for _, captureMap := range []map[bool]map[ltl.Token]struct{}{c.caps, oc.caps} {
		for matchingState := range captureMap {
			if captureMap[matchingState] != nil {
				for tok := range captureMap[matchingState] {
					ret.Capture(matchingState, tok)
				}
			}
		}
	}

	return ret
}

// Not returns a new Capture in which the captured tokens' matching states are
// inverted.
func (c *Captures) Not() *Captures {
	if c == nil {
		return nil
	}
	ret := New()
	ret.caps[true] = c.caps[false]
	ret.caps[false] = c.caps[true]
	return ret
}

// Reducible returns true if the receiver contains no captured tokens.
func (c *Captures) Reducible() bool {
	return c == nil || (len(c.caps[true]) == 0 && len(c.caps[false]) == 0)
}
