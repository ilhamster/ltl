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

// Package signals defines an ltl.Token type and matcher for multi-channel
// boolean signals.
package signals

import (
	"errors"
	"fmt"
	"ltl/pkg/ltl"
	"strings"
)

// signals is the basic type for both tokens and matchers.
type signals map[string]bool

func (sig signals) String() string {
	var ret []string
	for k, v := range sig {
		ret = append(ret, fmt.Sprintf("%s:%t", k, v))
	}
	return strings.Join(ret, ", ")
}

func newSignal(names ...string) signals {
	sig := signals{}
	for _, name := range names {
		if len(name) == 0 {
			continue
		}
		val := true
		if name[0] == '!' {
			name = name[1:]
			val = false
			if len(name) == 0 {
				continue
			}
		}
		sig[name] = val
	}
	return sig
}

// SignalToken is a Token type for multi-channel boolean signals.
type SignalToken signals

// NewToken returns a new Token containing a number of named boolean signals.
// Signals are specified by string name, which may be prefixed with '!' to
// indicate inversion.
func NewToken(names ...string) SignalToken {
	return SignalToken(newSignal(names...))
}

func (sigt SignalToken) String() string {
	return fmt.Sprintf("T %s", signals(sigt))
}

// EOI returns false for all SignalTokens.
func (sigt SignalToken) EOI() bool {
	return false
}

type signalMatcher signals

func (sigm signalMatcher) String() string {
	return fmt.Sprintf("M %s", signals(sigm))
}

func (sigm signalMatcher) Children() []ltl.Operator {
	return nil
}

func (sigm signalMatcher) Match(t ltl.Token) (ltl.Operator, ltl.Environment) {
	sigt, ok := t.(SignalToken)
	if !ok {
		return nil, ltl.ErrEnv(errors.New("not a stok"))
	}
	for k, v := range sigm {
		if tv, ok := sigt[k]; !ok || tv != v {
			return nil, ltl.NotMatching
		}
	}
	return nil, ltl.Matching
}

func (sigm signalMatcher) Reducible() bool {
	return true
}

// NewMatcher returns a new matcher matching a number of named boolean signals.
// Signals are specified by string name, which may be prefixed with '!' to
// indicate inversion.  On a Match, the resulting environment matches if the
// Token is a SignalToken containing all the signals specified by the matcher,
// with the same inversion status.
func NewMatcher(names ...string) ltl.Operator {
	return signalMatcher(newSignal(names...))
}
