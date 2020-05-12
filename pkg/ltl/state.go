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

package ltl

// State is the most basic Environment, conveying only matching status.
type State bool

const (
	NotMatching State = false
	Matching    State = true
)

func (s State) String() string {
	if s {
		return "Matching"
	}
	return "NotMatching"
}

func (s State) And(env Environment) Environment {
	if env.Err() != nil {
		return env
	}
	if _, ok := env.(State); !ok {
		// If the argument is not a State, it may have additional state and a
		// more complex And method, so it should be the receiver.
		return env.And(s)
	}
	if s {
		return env
	}
	return NotMatching
}

func (s State) Or(env Environment) Environment {
	if errEnv := EitherErroring(s, env); errEnv != nil {
		return errEnv
	}
	if _, ok := env.(State); !ok {
		// If the argument is not a State, it may have additional state and a
		// more complex Or method, so it should be the receiver.
		return env.Or(s)
	}
	if env.Matching() {
		return env
	}
	return s
}

func (s State) Not() Environment {
	if s {
		return NotMatching
	}
	return Matching
}

func (s State) Matching() bool {
	return bool(s)
}

func (s State) Err() error {
	return nil
}

func (s State) Reducible() bool {
	return true
}
