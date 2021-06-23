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

// Package tags provides a type that associates tags with matching Environments.
package tags

import (
	"fmt"
	"sort"
	"strings"
)

// Tags contains a set of distinct Tag objects.  A nil Tags is treated as
// empty.
type Tags struct {
	// t is stored in increasing comparison order.
	t []Tag
}

// newSorted returns a new Tags with the provided Tag slice.  It assumes the
// arguments are presented in (increasing) sorted order.
func newSorted(tags ...Tag) *Tags {
	if len(tags) == 0 {
		return nil
	}
	return &Tags{
		t: tags,
	}
}

// New returns a new Tags containing the provided Tag arguments.
func New(tags ...Tag) *Tags {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Compare(tags[j]) < 0
	})
	return newSorted(tags...)
}

func (t *Tags) Tags() []Tag {
	if t == nil {
		return nil
	}
	return t.t
}

func (t *Tags) String() string {
	ret := make([]string, 0, t.Length())
	for _, tag := range t.Tags() {
		ret = append(ret, tag.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(ret, ", "))
}

func (t *Tags) Length() int {
	return len(t.Tags())
}

func (t *Tags) Union(ot *Tags) *Tags {
	if t.Length() == 0 {
		return ot
	}
	if ot.Length() == 0 || t.Eq(ot) {
		return t
	}
	ret := make([]Tag, 0, t.Length()+ot.Length())
	tIdx, otIdx := 0, 0
	for tIdx < t.Length() && otIdx < ot.Length() {
		tT, otT := t.Tags()[tIdx], ot.Tags()[otIdx]
		cmp := tT.Compare(otT)
		if cmp < 0 {
			ret = append(ret, tT)
			tIdx++
		} else if cmp > 0 {
			ret = append(ret, otT)
			otIdx++
		} else {
			ret = append(ret, tT)
			tIdx++
			otIdx++
		}
	}
	ret = append(ret, t.Tags()[tIdx:]...)
	ret = append(ret, ot.Tags()[otIdx:]...)
	return newSorted(ret...)
}

func (t *Tags) Eq(ot *Tags) bool {
	if t.Length() != ot.Length() {
		return false
	}
	for idx := 0; idx < t.Length(); idx++ {
		if t.Tags()[idx].Compare(ot.Tags()[idx]) != 0 {
			return false
		}
	}
	return true
}
