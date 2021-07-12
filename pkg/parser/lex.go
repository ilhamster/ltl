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

// Package parser includes a basic parser for LTL expressions.
package parser

import (
	"bufio"
	"fmt"
	"io"
	"ltl/pkg/ltl"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type prefixNode struct {
	value    int
	children map[string]*prefixNode
}

type token struct {
	str   string
	value int
}

func newPrefixTree(tokenMap map[string]int) (*prefixNode, error) {
	var tokens []token
	for str, value := range tokenMap {
		tokens = append(tokens, token{str, value})
	}
	sort.Slice(tokens, func(a, b int) bool {
		return strings.Compare(tokens[a].str, tokens[b].str) < 0
	})
	return buildPrefixTree("", tokens)
}

func buildPrefixTree(rootPrefix string, tokens []token) (*prefixNode, error) {
	p := &prefixNode{
		value:    yyErrCode,
		children: map[string]*prefixNode{},
	}
	var thisPrefix string
	var theseTokens []token
	var err error
	for _, tok := range tokens {
		if len(tok.str) == 0 {
			if p.value != yyErrCode {
				return nil, fmt.Errorf("token %s has duplicate definitions", rootPrefix)
			}
			p.value = tok.value
			continue
		}
		if thisPrefix != "" && !strings.HasPrefix(tok.str, thisPrefix) {
			p.children[thisPrefix], err = buildPrefixTree(rootPrefix+thisPrefix, theseTokens)
			if err != nil {
				return nil, err
			}
			thisPrefix = ""
			theseTokens = nil
		}
		if thisPrefix == "" {
			thisPrefix = string(tok.str[0])
		}
		theseTokens = append(theseTokens, token{
			tok.str[1:], tok.value,
		})
	}
	if thisPrefix != "" {
		p.children[thisPrefix], err = buildPrefixTree(rootPrefix+thisPrefix, theseTokens)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

// advance applies the provided rune to the receiver, returning the new
// prefixNode.  If a nil prefixNode is returned, the
func (p *prefixNode) advance(r rune) *prefixNode {
	return p.children[string(r)]
}

// lookup returns the parser token associated with a string in this prefixTree,
// or yyErrCode if the string is not in the prefixTree.
func (p *prefixNode) lookup(str string) int {
	if len(str) == 0 {
		return p.value
	}
	newP, ok := p.children[string(str[0])]
	if !ok {
		return yyErrCode
	}
	return newP.lookup(str[1:])
}

var (
	// DefaultTokens is a default mapping of token strings to token values.
	DefaultTokens = map[string]int{
		"AND":        AND,
		"LIMIT":      LIMIT,
		"EVENTUALLY": EVENTUALLY,
		"NEXT":       NEXT,
		"NOT":        NOT,
		"OR":         OR,
		"SEQUENCE":   SEQUENCE,
		"THEN":       THEN,
		"UNTIL":      UNTIL,
		"RELEASE":    RELEASE,
		"GLOBALLY":   GLOBALLY,
	}
	// OpenParen is a default open-parenthesis symbol.
	OpenParen rune = '('
	// CloseParen is a default close-parenthesis symbol.
	CloseParen rune = ')'
	// OpenBracket is a default open-bracket symbol.  In this lexer, brackets
	// enclose text to be sent to a 'matcher' (a terminal ltl.Operator).  This
	// text may itself contain brackets, but they must be balanced.
	OpenBracket rune = '['
	// CloseBracket is a default close-bracket symbol.
	CloseBracket rune = ']'
)

// Lexer is a lexer used by ParseLTL to parse expression strings into LTL
// Operations.
type Lexer struct {
	r                 *bufio.Reader
	matcherGenerator  func(string) (ltl.Operator, error)
	rootPrefixTree    *prefixNode
	currentPrefixTree *prefixNode
	offset            int
	op                ltl.Operator
	// yyLexer.Lex returns only an int, not also an error.  So, to signal a
	// lexing error, Lexer::Lex must set an error (to be retrieved later with
	// Lexer::Error).  If Lex sets a non-nil error, it should immediately return
	// yyErrCode.
	err error
}

// NewLexer returns a new lexer, using the token set in tokens, and the
// matcherGenerator function to convert matcher strings to Operators.
func NewLexer(tokens map[string]int, matcherGenerator func(string) (ltl.Operator, error), r *bufio.Reader) (*Lexer, error) {
	p, err := newPrefixTree(tokens)
	if err != nil {
		return nil, err
	}
	return &Lexer{
		r:                 r,
		matcherGenerator:  matcherGenerator,
		rootPrefixTree:    p,
		currentPrefixTree: p,
		offset:            0,
	}, nil
}

func (l *Lexer) Lex(lvalue *yySymType) int {
	var r rune
	var c int
	var err error
	// Consume runes until an EOF, error, or non-whitespace rune is
	// encountered.
	for {
		r, c, err = l.r.ReadRune()
		l.offset += c
		if err == io.EOF {
			return -1
		}
		if err != nil {
			l.err = fmt.Errorf("read error at offset %d: %s", l.offset, err)
			return yyErrCode
		}
		if !unicode.Is(unicode.White_Space, r) {
			break
		}
	}
	switch {
	case r == OpenParen:
		return LPAREN
	case r == CloseParen:
		return RPAREN
	case r == OpenBracket:
		matcherStr := ""
		bracketDepth := 1
		for {
			if bracketDepth == 0 {
				break
			}
			r, c, err = l.r.ReadRune()
			l.offset += c
			if err == io.EOF {
				l.err = fmt.Errorf("unexpected EOF at offset %d", l.offset)
				return yyErrCode
			}
			switch r {
			case OpenBracket:
				bracketDepth++
			case CloseBracket:
				bracketDepth--
			default:
				matcherStr += string(r)
			}
		}
		op, err := l.matcherGenerator(matcherStr)
		if err != nil {
			l.err = fmt.Errorf("failed to create matcher ending at offset %d: %s", l.offset, err)
			return yyErrCode
		}
		lvalue.op = op
		return MATCHER
	case r == CloseBracket:
		l.err = fmt.Errorf("unexpected '%c' at offset %d", CloseBracket, l.offset)
		return yyErrCode
	case unicode.IsDigit(r):
		l.r.UnreadRune()
		var num string
		for {
			r, c, err := l.r.ReadRune()
			l.offset += c
			if err != nil && err != io.EOF {
				l.err = fmt.Errorf("read error at offset %d: %s", l.offset, err)
				return yyErrCode
			}
			if !unicode.IsDigit(r) {
				l.r.UnreadRune()
			}
			if err == io.EOF || !unicode.IsDigit(r) {
				lvalue.num, err = strconv.ParseInt(num, 10, 64)
				if err != nil {
					l.err = fmt.Errorf("failed to parse number %s: %s", num, err)
					return yyErrCode
				}
				return NUM
			}
			num = num + string(r)
		}
	default:
		l.r.UnreadRune()
		for {
			r, c, err := l.r.ReadRune()
			l.offset += c
			if err != nil && err != io.EOF {
				l.err = fmt.Errorf("read error at offset %d: %s", l.offset, err)
				return yyErrCode
			}
			if err == io.EOF || unicode.Is(unicode.White_Space, r) {
				ret := l.currentPrefixTree.value
				l.currentPrefixTree = l.rootPrefixTree
				return ret
			}
			next := l.currentPrefixTree.advance(r)
			if next == nil {
				l.err = fmt.Errorf("lexing error at offset %d", l.offset)
				return yyErrCode
			}
			l.currentPrefixTree = next
		}
	}
}

func (l *Lexer) Error(e string) {
	l.err = fmt.Errorf("parse error at offset %d: %s", l.offset, e)
}
