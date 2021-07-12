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

type errEnv struct {
	error
}

// ErrEnv returns an Erroring Environment associating the specified error.
func ErrEnv(err error) Environment {
	return errEnv{err}
}

func (e errEnv) String() string {
	return e.Err().Error()
}

func (e errEnv) And(env Environment) Environment {
	return e
}

func (e errEnv) Or(env Environment) Environment {
	return e
}

func (e errEnv) Not() Environment {
	return e
}

func (e errEnv) Clone() Environment {
	return e
}

func (e errEnv) Matching() bool {
	return false
}

func (e errEnv) Err() error {
	return e.error
}

func (e errEnv) Reducible() bool {
	return true
}
