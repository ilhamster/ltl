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
)

// BoundValue represents a value bound to a string key.  Bound values must be
// comparable, and bound values comparing as equal must actually be equal.
type BoundValue interface {
	fmt.Stringer
	// Type returns the value type of this BoundValue.  BoundValues of different
	// value types must *always* return distinct Type(), so it is good practice
	// to return the value type name here.
	Type() string
	// CompareValues returns <0, 0, or >0 if the receiver's value compares less
	// than, equal, or greater than the argument's value, respectively.  If the
	// two are incomparable, due to different value types, an error should be
	// returned.
	CompareValues(obv BoundValue) (int, error)
	// Key returns the key to which this BoundValue binds.
	Key() string
}
