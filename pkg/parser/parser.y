// goyacc -o parser/parser.go parser/parser.y
%{
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
    "github.com/ilhamster/ltl/pkg/ltl"
    ops "github.com/ilhamster/ltl/pkg/operators"
)
%}

// yySymType
%union{
    op ltl.Operator
    num int64
}

%type <op> line expr

%token <op> MATCHER

%token <num> NUM

%token LPAREN RPAREN

%nonassoc LIMIT
%nonassoc GLOBALLY
%nonassoc EVENTUALLY
%left UNTIL RELEASE
%left THEN SEQUENCE
%left OR AND
%left NEXT NOT

%start line 

%%

line : expr                { setOp(yylex, $1) }
     ;

expr : LPAREN expr RPAREN  { $$ = $2 }
     | MATCHER             { $$ = $1 }
     | NOT expr            { $$ = ops.Not($2) }
     | NEXT expr           { $$ = ops.Next($2) }
     | EVENTUALLY expr     { $$ = ops.Eventually($2) }
     | GLOBALLY expr       { $$ = ops.Globally($2) }
     | expr LIMIT NUM      { $$ = ops.Limit($3, $1) }
     | expr OR expr        { $$ = ops.Or($1, $3) }
     | expr AND expr       { $$ = ops.And($1, $3) }
     | expr UNTIL expr     { $$ = ops.Until($1, $3) }
     | expr RELEASE expr   { $$ = ops.Release($1, $3) }
     | expr THEN expr      { $$ = ops.Then($1, $3) }
     ;

%%

func setOp(l yyLexer, op ltl.Operator) {
    l.(*Lexer).op = op
}

type yyLex struct {
    s string
    pos int
}

// ParseLTL parses an expression, lexed by the provided Lexer, into an LTL
// Operator.
func ParseLTL(l *Lexer) (ltl.Operator, error) {
    yyErrorVerbose = true
    p := &yyParserImpl{}
    p.Parse(l)
    return l.op, l.err
}
