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
	"bufio"
	"github.com/ilhamster/ltl/examples/stringmatcher"
	"github.com/ilhamster/ltl/pkg/ltl"
	ops "github.com/ilhamster/ltl/pkg/operators"
	"strings"
	"testing"
)

func parse(s string) (ltl.Operator, int, int, error) {
	l, err := NewLexer(DefaultTokens,
		stringmatcher.Generator(),
		bufio.NewReader(strings.NewReader(s)))
	if err != nil {
		return nil, 0, 0, err
	}
	op, err := ParseLTL(l)
	return op, l.LastTokenStartOffset(), l.Offset(), err
}

func TestParser(t *testing.T) {
	tests := []struct {
		description              string
		input                    string
		wantErr                  bool
		wantLastTokenStartOffset int
		wantOffset               int
	}{{
		"normal parsing",
		"[a] THEN [b]",
		false,
		9,
		12, // After the expression
	}, {
		"parse error",
		"[a] [b] AND [c]",
		true,
		4,
		7, // After the [b]
	}, {
		"matcher error",
		"[$] AND [c]",
		true,
		0,
		3, // After the [$]
	}, {
		"lexing error",
		"[a] WHEREUPON [b]",
		true,
		4,
		5, // After the 'W'
	}}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			_, gotLastTokenStartOffset, gotOffset, err := parse(test.input)
			if err != nil && !test.wantErr {
				t.Fatalf("Parse expected no error, but got %s", err)
			}
			if err == nil && test.wantErr {
				t.Fatalf("Parse expected an error, but got none")
			}
			if gotLastTokenStartOffset != test.wantLastTokenStartOffset {
				t.Errorf("Last token start offset was %d, but wanted %d", gotLastTokenStartOffset, test.wantLastTokenStartOffset)
			}
			if gotOffset != test.wantOffset {
				t.Errorf("Reached offset %d, but wanted %d", gotOffset, test.wantOffset)
			}
		})
	}
}

// Also tests precedence.
func TestParsingAsString(t *testing.T) {
	tests := []struct {
		input     string
		wantOpStr string
	}{{
		"[a] THEN [b] ",
		"THEN([a],[b])",
	}, {
		"(EVENTUALLY [a]) LIMIT 10 ",
		"LIMIT(10)(EVENTUALLY([a]))",
	}, {
		"EVENTUALLY [a] THEN [b]",
		"EVENTUALLY(THEN([a],[b]))",
	}, {
		// But seriously, use parens.
		"[a] UNTIL [b] THEN [c]",
		"UNTIL([a],THEN([b],[c]))",
	}, {
		"[a] THEN [b] UNTIL [c]",
		"UNTIL(THEN([a],[b]),[c])",
	}, {
		"[a] THEN EVENTUALLY [b] THEN [c]",
		"THEN([a],EVENTUALLY(THEN([b],[c])))",
	}, {
		"[a] THEN NOT [b]",
		"THEN([a],NOT([b]))",
	}, {
		"NOT [a] THEN [b]",
		"THEN(NOT([a]),[b])",
	}, {
		"NOT [a] AND [b]",
		"AND(NOT([a]),[b])",
	}}
	for _, test := range tests {
		op, _, _, err := parse(test.input)
		if err != nil {
			t.Fatalf("Failed to parse: %s", err)
		}
		if strings.Compare(ops.PrettyPrint(op, ops.Inline()), test.wantOpStr) != 0 {
			t.Fatalf("Wanted parsed operation %s, got %s", test.wantOpStr, ops.PrettyPrint(op, ops.Inline()))
		}
	}
}
