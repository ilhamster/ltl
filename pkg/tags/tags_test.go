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
	"testing"
)

func TestIndexTags(t *testing.T) {
	for _, test := range []struct {
		a, b      *Tags
		wantUnion *Tags
	}{{
		New(Index(1), Index(2), Index(3)),
		New(Index(3), Index(4), Index(5)),
		New(Index(1), Index(2), Index(3), Index(4), Index(5)),
	}, {
		New(Index(1)),
		nil,
		New(Index(1)),
	}, {
		New(Index(1), Index(2)),
		New(Index(3)),
		New(Index(1), Index(2), Index(3)),
	}} {
		t.Run(fmt.Sprintf("%s U %s", test.a, test.b), func(t *testing.T) {
			gotUnion := test.a.Union(test.b)
			if !test.wantUnion.Eq(gotUnion) {
				t.Fatalf("Union() = %s, want %s", gotUnion, test.wantUnion)
			}
		})
	}
}
