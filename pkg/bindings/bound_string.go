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
	"strings"
)

// BoundString is a single string bound to a key.
type BoundString struct {
	key, value string
}

// String returns an string value bound to a key.
func String(key, value string) *BoundString {
	return &BoundString{
		key:   key,
		value: value,
	}
}

// Type returns 'string' for BoundStrings.
func (bs *BoundString) Type() string {
	return "string"
}

// CompareValues compares the receiver and argument.
func (bs *BoundString) CompareValues(obv BoundValue) (int, error) {
	obs, ok := obv.(*BoundString)
	if !ok {
		return 0, fmt.Errorf("BoundValue %s had type %T, expected *BoundString", obv, obv)
	}
	return strings.Compare(bs.value, obs.value), nil
}

// Key returns the key of the receiver.
func (bs *BoundString) Key() string {
	return bs.key
}

func (bs *BoundString) String() string {
	return fmt.Sprintf("%s:%s", bs.key, bs.value)
}
