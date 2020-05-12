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

// boundString is a single string bound to a key.
type boundString struct {
	key, value string
}

// Int returns an string value bound to a key.
func String(key, value string) *boundString {
	return &boundString{
		key:   key,
		value: value,
	}
}

func (bs *boundString) Type() string {
	return "string"
}

func (bs *boundString) CompareValues(obv BoundValue) (int, error) {
	obs, ok := obv.(*boundString)
	if !ok {
		return 0, fmt.Errorf("BoundValue %s had type %T, expected *boundString", obv, obv)
	}
	return strings.Compare(bs.value, obs.value), nil
}

func (bs *boundString) Key() string {
	return bs.key
}

func (bs *boundString) String() string {
	return fmt.Sprintf("%s:%s", bs.key, bs.value)
}
