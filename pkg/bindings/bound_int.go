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

// BoundInt is a single int bound to a key.
type BoundInt struct {
	key   string
	value int
}

// Int returns an integer value bound to a key.
func Int(key string, value int) *BoundInt {
	return &BoundInt{
		key:   key,
		value: value,
	}
}

// Type returns 'int' for BoundInts.
func (bi *BoundInt) Type() string {
	return "int"
}

// CompareValues compares the receiver and argument.
func (bi *BoundInt) CompareValues(obv BoundValue) (int, error) {
	obi, ok := obv.(*BoundInt)
	if !ok {
		return 0, fmt.Errorf("BoundValue %s had type %T, expected *BoundInt", obv, obv)
	}
	return bi.value - obi.value, nil
}

// Key returns the key of the receiver.
func (bi *BoundInt) Key() string {
	return bi.key
}

func (bi *BoundInt) String() string {
	return fmt.Sprintf("%s:%d", bi.key, bi.value)
}
