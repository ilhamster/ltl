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

// Package stringmatcher provides a terminal string-matching Operator.  This
// Operator consumes rune tokens until its string is fully matched, returning
// early, without matching, on a difference.  It also supports binding and
// referencing token values.
package stringmatcher

import (
	"errors"
	"fmt"
	rt "ltl/examples/runetoken"
	"ltl/pkg/binder"
	be "ltl/pkg/bindingenvironment"
	"ltl/pkg/bindings"
	"ltl/pkg/ltl"
	"strings"
)

type config struct {
	caseSensitive bool
	capture       bool
}

type Option func(c *config)

// Capture specifies whether matching tokens should be captured in the
// Environment.
func Capture(capture bool) Option {
	return func(c *config) {
		c.capture = capture
	}
}

// CaseSensitive specifies whether string matches are case sensitive.  Defaults
// to false.
func CaseSensitive(caseSensitive bool) Option {
	return func(c *config) {
		c.caseSensitive = caseSensitive
	}
}

type stringMatcher struct {
	s string
	c *config
}

func new(s string, c *config) *stringMatcher {
	if !c.caseSensitive {
		s = strings.ToLower(s)
	}
	return &stringMatcher{s: s, c: c}
}

// New returns a new ltl.Operator that matches the provided string under the
// provided Options.  Strings may be matched piecemeal; if, on a Match, the
// provided Token is a prefix of the string to be matched, the returned Operator
// will match the remaining suffix of the original string.
func New(s string, opts ...Option) *stringMatcher {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return new(s, c)
}

func (sm *stringMatcher) matchInternal(rtok *rt.RuneToken) (ltl.Operator, ltl.Environment) {
	if len(sm.s) == 0 || rtok.EOI() {
		return nil, be.New(be.Matching(false))
	}
	matching := false
	var rem string
	if sm.s[0] == '.' {
		rem = sm.s[1:]
		matching = true
	} else {
		val := string(rtok.Value())
		if !sm.c.caseSensitive {
			val = strings.ToLower(val)
		}
		if strings.HasPrefix(sm.s, string(val)) {
			rem = strings.TrimPrefix(sm.s, string(val))
			matching = len(rem) == 0
		}
	}
	opts := []be.Option{be.Matching(matching)}
	if sm.c.capture {
		opts = append(opts, be.Captured(rtok))
	}
	env := be.New(opts...)
	if len(rem) > 0 {
		return new(rem, sm.c), env
	}
	return nil, env
}

func (sm *stringMatcher) Match(tok ltl.Token) (ltl.Operator, ltl.Environment) {
	rtok, ok := tok.(*rt.RuneToken)
	if !ok {
		return nil, ltl.ErrEnv(errors.New("expected *rt.RuneToken"))
	}
	return sm.matchInternal(rtok)
}

func (sm stringMatcher) String() string {
	return fmt.Sprintf("[%s]", sm.s)
}

func (sm *stringMatcher) Reducible() bool {
	return true
}

// Generator returns a generator function producing string matchers with the
// specified options.  The returned function accepts a string and returns a
// matcher for that string (and possibly an error).
func Generator(opts ...Option) func(s string) (ltl.Operator, error) {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	bindingBuilder := binder.NewBuilder(c.capture, func(name string, tok ltl.Token) (*bindings.Bindings, error) {
		rtok, ok := tok.(*rt.RuneToken)
		if !ok {
			return nil, fmt.Errorf("failed to make Bindings: require *rt.RuneToken")
		}
		bs, err := bindings.New(bindings.String(name, string(rtok.Value())))
		return bs, err
	})

	return func(s string) (ltl.Operator, error) {
		if strings.HasPrefix(s, "$") {
			s = strings.TrimPrefix(s, "$")
			if strings.HasSuffix(s, "<-") {
				s = strings.TrimSuffix(s, "<-")
				s = strings.TrimSpace(s)
				if len(s) == 0 {
					return nil, fmt.Errorf("failed to make binding: no name specified")
				}
				return bindingBuilder.Bind(s), nil
			}
			s = strings.TrimSpace(s)
			if len(s) == 0 {
				return nil, fmt.Errorf("failed to make reference: no name specified")
			}
			return bindingBuilder.Reference(s), nil
		}
		return new(s, c), nil
	}
}
