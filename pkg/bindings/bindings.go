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

// Package bindings provides an interface and helper functions for types that
// bind values (of arbitrary type) to string names.
package bindings

import (
	"fmt"
	"sort"
	"strings"
)

// Bindings is a set of BoundValues.  A nil Bindings is treated as empty.
type Bindings struct {
	// b is stored by increasing key.
	b []BoundValue
}

func (b *Bindings) bindings() []BoundValue {
	if b == nil {
		return nil
	}
	return b.b
}

// newSorted returns a new Bindings with the provided BoundValue slice.  It
// assumes the arguments are presented in (increasing) sorted order.
func newSorted(bvs ...BoundValue) *Bindings {
	if len(bvs) == 0 {
		return nil
	}
	return &Bindings{
		b: bvs,
	}
}

// New returns a new Bindings with the provided BoundValues.
func New(bvs ...BoundValue) (*Bindings, error) {
	var err error
	sort.Slice(bvs, func(i, j int) bool {
		cmp := strings.Compare(bvs[i].Key(), bvs[j].Key())
		if cmp == 0 {
			err = fmt.Errorf("Key conflict between %s and %s", bvs[i], bvs[j])
		}
		return cmp < 0
	})
	return newSorted(bvs...), err
}

func (b *Bindings) String() string {
	ret := make([]string, 0, len(b.bindings()))
	for _, bv := range b.bindings() {
		ret = append(ret, bv.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(ret, ", "))
}

// Length returns the number of bound names in the receiver.
func (b *Bindings) Length() int {
	return len(b.bindings())
}

// Combine combines the receiver and argument Bindings, adding each key and
// value in the argument into the receiver.  If the Binding types are
// incompatible or if the same key exists in both combined Bindings, Combine
// should return an error.
func (b *Bindings) Combine(ob *Bindings) (*Bindings, error) {
	// Performance: if b is empty, or it's the same as ob, we can just return
	// ob.
	if b.Length() == 0 || b.Eq(ob) {
		return ob, nil
	}
	// Performance: if ob is empty, we can just return b.
	if ob.Length() == 0 {
		return b, nil
	}
	ret := make([]BoundValue, 0, b.Length()+ob.Length())
	bIdx, obIdx := 0, 0
	for bIdx < b.Length() && obIdx < ob.Length() {
		bBV, oBV := b.bindings()[bIdx], ob.bindings()[obIdx]
		cmp := strings.Compare(bBV.Key(), oBV.Key())
		if cmp < 0 {
			ret = append(ret, bBV)
			bIdx++
		}
		if cmp > 0 {
			ret = append(ret, oBV)
			obIdx++
		}
		if cmp == 0 {
			if cmp, err := bBV.CompareValues(oBV); err != nil {
				return nil, err
			} else if cmp != 0 {
				return nil, fmt.Errorf("Key %s conflicts in %s and %s", bBV.Key(), b, ob)
			}
			ret = append(ret, bBV)
			bIdx++
			obIdx++
		}
	}
	ret = append(ret, b.bindings()[bIdx:]...)
	ret = append(ret, ob.bindings()[obIdx:]...)
	return newSorted(ret...), nil
}

// Satisfy returns the relative complement of the argument in the receiver: that
// is, a copy of the receiver with all keys also present in the argument (and
// with the same value) removed.  It returns true if the receiver could be
// satisfied by the argument: if every bound name present in both the receiver
// and the argument binds to the same value in both.  Note that a return value
// of true does not imply either that the returned Bindings is empty.
func (b *Bindings) Satisfy(ob *Bindings) (*Bindings, bool) {
	// Performance: if either is empty, we can just return the receiver.
	if b.Length() == 0 || ob.Length() == 0 {
		return b, true
	}
	// Performance: if the argument and receiver are equal, the result is a
	// fully-satisfied (that is, empty) Bindings.
	if b.Eq(ob) {
		return nil, true
	}
	ret := make([]BoundValue, 0, b.Length())
	bIdx, obIdx := 0, 0
	for bIdx < b.Length() && obIdx < ob.Length() {
		bBV, oBV := b.bindings()[bIdx], ob.bindings()[obIdx]
		cmp := strings.Compare(bBV.Key(), oBV.Key())
		if cmp < 0 {
			ret = append(ret, bBV)
			bIdx++
		}
		if cmp > 0 {
			obIdx++
		}
		if cmp == 0 {
			if cmp, err := bBV.CompareValues(oBV); err != nil || cmp != 0 {
				return nil, false
			}
			bIdx++
			obIdx++
		}
	}
	ret = append(ret, b.bindings()[bIdx:]...)
	return newSorted(ret...), true
}

// Eq compares the receiver and the argument.  Bindings are identical iff they
// are of the same type and have the same keys bound to the same values.
func (b *Bindings) Eq(ob *Bindings) bool {
	if b.Length() != ob.Length() {
		return false
	}
	for idx := 0; idx < b.Length(); idx++ {
		bBV, oBV := b.bindings()[idx], ob.bindings()[idx]
		if bBV.Key() != oBV.Key() {
			return false
		}
		if cmp, err := bBV.CompareValues(oBV); cmp != 0 || err != nil {
			return false
		}
	}
	return true
}

// Keys returns the set of bound names in the receiver.
func (b *Bindings) Keys() map[string]struct{} {
	ret := map[string]struct{}{}
	for _, bv := range b.bindings() {
		ret[bv.Key()] = struct{}{}
	}
	return ret
}
