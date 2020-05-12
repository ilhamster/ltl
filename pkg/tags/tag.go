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

package tags

import (
	"fmt"
)

// Tag represents a single tag value associated with a matching Environment.
type Tag interface {
	fmt.Stringer
	// Type returns the type of this Tag, as a string.  Comparable tags with the
	// same semantics should return the same Type, incomparable ones or ones
	// with different semantics should return distinct Types.  Unlike
	// BoundValues, it is *not* best practice to return the value type here:
	// tags representing input indices and tags representing some enum may both
	// be int, but have distinct meanings.
	Type() string
	// Compare returns <0, 0, or >0 if the receiver compares less than, equal
	// to, or greater than the argument.  If the two have different Types,
	// those Types should be compared; otherwise the tag values are compared.
	Compare(ot Tag) int
}
