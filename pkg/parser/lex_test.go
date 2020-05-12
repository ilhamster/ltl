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

package parser

import (
	"testing"
)

func TestPrefixTree(t *testing.T) {
	tests := []struct {
		description string
		tokenMap    map[string]int
		inputs      map[string]int
	}{{
		"one-to-one",
		map[string]int{
			"AND":        AND,
			"EVENTUALLY": EVENTUALLY,
			"NEXT":       NEXT,
			"NOT":        NOT,
			"OR":         OR,
			"SEQUENCE":   SEQUENCE,
			"THEN":       THEN,
			"UNTIL":      UNTIL,
		},
		map[string]int{
			"AND":      AND,
			"WHERE":    yyErrCode,
			"SEQUENCE": SEQUENCE,
		},
	}, {
		"prefix",
		map[string]int{
			"AND":     AND,
			"ANDTHEN": THEN,
			"N":       NOT,
			"NEXT":    NEXT,
		},
		map[string]int{
			"AND":     AND,
			"ANDTHEN": THEN,
			"N":       NOT,
			"NE":      yyErrCode,
		},
	}, {
		"multiple spellings",
		map[string]int{
			"AND":        AND,
			"&":          AND,
			"OR":         OR,
			"|":          OR,
			"EVENTUALLY": EVENTUALLY,
			"E":          EVENTUALLY,
		},
		map[string]int{
			"AND":        AND,
			"&":          AND,
			"OR":         OR,
			"|":          OR,
			"EVENTUALLY": EVENTUALLY,
			"E":          EVENTUALLY,
		},
	}}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			p, err := newPrefixTree(test.tokenMap)
			if err != nil {
				t.Fatalf("newPrefixTree yielded error %s, wanted none", err)
			}
			for input, wantOutput := range test.inputs {
				gotOutput := p.lookup(input)
				if gotOutput != wantOutput {
					t.Errorf("lookup() = %d, want %d", gotOutput, wantOutput)
				}
			}
		})
	}
}
