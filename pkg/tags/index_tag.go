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
	"strings"
)

// indexTag tags an Environment with an index into the input string.
type indexTag struct {
	index int
}

func (it *indexTag) String() string {
	return fmt.Sprintf("%d", it.index)
}

func (it *indexTag) Type() string {
	return "index"
}

func (it *indexTag) Compare(ot Tag) int {
	if oit, ok := ot.(*indexTag); ok {
		if it.index < oit.index {
			return -1
		}
		if it.index > oit.index {
			return 1
		}
		return 0
	}
	return strings.Compare(it.Type(), ot.Type())
}

func IndexOf(t Tag) (int, bool) {
	if it, ok := t.(*indexTag); ok {
		return it.index, true
	}
	return 0, false
}

// Index returns a new index tag with the provided Index.
func Index(index int) *indexTag {
	return &indexTag{
		index: index,
	}
}
