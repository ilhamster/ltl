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

// boundInt is a single int bound to a key.
type boundInt struct {
	key   string
	value int
}

// Int returns an integer value bound to a key.
func Int(key string, value int) *boundInt {
	return &boundInt{
		key:   key,
		value: value,
	}
}

func (bi *boundInt) Type() string {
	return "int"
}

func (bi *boundInt) CompareValues(obv BoundValue) (int, error) {
	obi, ok := obv.(*boundInt)
	if !ok {
		return 0, fmt.Errorf("BoundValue %s had type %T, expected *boundInt", obv, obv)
	}
	return bi.value - obi.value, nil
}

func (bi *boundInt) Key() string {
	return bi.key
}

func (bi *boundInt) String() string {
	return fmt.Sprintf("%s:%d", bi.key, bi.value)
}
