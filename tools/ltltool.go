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

// Binary ltltool is a command-line tool for testing string-binding LTL queries.
package main

import (
	"bufio"
	"flag"
	"fmt"
	rt "github.com/ilhamster/ltl/examples/runetoken"
	smatch "github.com/ilhamster/ltl/examples/stringmatcher"
	be "github.com/ilhamster/ltl/pkg/bindingenvironment"
	"github.com/ilhamster/ltl/pkg/ltl"
	ops "github.com/ilhamster/ltl/pkg/operators"
	"github.com/ilhamster/ltl/pkg/parser"
	"io"
	"log"
	"os"
	"strings"
)

var (
	inFilename = flag.String("filename", "", "A file containing commands to run before entering interactive mode.")
)

type ltlif struct {
	op                                ltl.Operator
	expEnv, expMatches, expOp, expTok bool
	capture                           bool
	caseSensitive                     bool
}

func newIf() *ltlif {
	return &ltlif{}
}

func (lif *ltlif) parse(s string) (ltl.Operator, error) {
	l, err := parser.NewLexer(parser.DefaultTokens,
		smatch.Generator(smatch.Capture(lif.capture), smatch.CaseSensitive(lif.caseSensitive)),
		bufio.NewReader(strings.NewReader(s)))
	if err != nil {
		return nil, err
	}
	return parser.ParseLTL(l)
}

func (lif *ltlif) setOp(expression string) {
	op, err := lif.parse(expression)
	if err != nil {
		fmt.Printf("Parse error: %s\n", err.Error())
		return
	}
	lif.op = op
	fmt.Printf("Operator set to: \n%s\n", ops.PrettyPrint(lif.op, ops.Prefix(" | ")))
	return
}

func (lif *ltlif) run(input string) {
	if lif.op == nil {
		fmt.Println("No operator is set, set one with 'op <expression>'")
		return
	}
	op := lif.op
	var env ltl.Environment
	reader := strings.NewReader(input)
	for index := 0; ; index++ {
		r, _, err := reader.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Read error: %s\n", err)
			return
		}
		tok := rt.New(r, index)
		if lif.expTok {
			fmt.Printf("Token %s\n", tok)
		}
		if lif.expOp {
			fmt.Println("Op:")
			ops.PrettyPrint(op, ops.Prefix("  : "))
		}
		if lif.expEnv {
			fmt.Println("Env:")
			be.PrettyPrint(env, "   : ")
		}
		if op == nil {
			fmt.Printf("Stopped parsing at token %d.\n  Last environment was ", index)
			if env.Matching() {
				fmt.Println("matching:")
			} else {
				fmt.Println("not matching:")
			}
			be.PrettyPrint(env)
			return
		}
		op, env = ltl.Match(op, tok)
		if lif.expMatches && env != nil && env.Matching() {
			fmt.Println("matching:")
			be.PrettyPrint(env)
		}
	}
	if env.Err() != nil {
		fmt.Printf("Environment reports error: %s\n", env.Err())
	}
	if env.Matching() {
		fmt.Println("Match!")
	} else {
		fmt.Println("No match")
	}
	be.PrettyPrint(env)
}

const (
	letterCase = "case"
	explain    = "explain"
	help       = "help"
	op         = "op"
	quit       = "quit"
	run        = "run"
	capture    = "capture"
)

func (lif *ltlif) do(in string) {
	in = strings.TrimSpace(in)
	if len(in) == 0 || in[0] == '#' {
		return
	}
	parts := strings.SplitN(in, " ", 2)
	cmd, remainder := parts[0], parts[1:]
	switch cmd {
	case op:
		if len(remainder) != 1 {
			break
		}
		lif.setOp(remainder[0])
		return
	case run:
		if len(remainder) != 1 {
			break
		}
		lif.run(remainder[0])
		return
	case explain:
		if len(remainder) != 1 {
			break
		}
		switch remainder[0] {
		case "all":
			lif.expEnv, lif.expMatches, lif.expOp, lif.expTok = true, true, true, true
			fmt.Println("Explaining tokens, operations, and environments.")
			return
		case "envs", "environments":
			lif.expEnv = true
			fmt.Println("Explaining environments.")
			return
		case "matches":
			lif.expMatches = true
			fmt.Println("Explaining matches.")
			return
		case "nothing":
			lif.expEnv, lif.expMatches, lif.expOp, lif.expTok = false, false, false, false
			fmt.Println("Not explaining tokens, operations, and environments.")
			return
		case "ops", "operations":
			lif.expOp = true
			fmt.Println("Explaining operations.")
			return
		case "toks", "tokens":
			lif.expTok = true
			fmt.Println("Explaining tokens.")
			return
		default:
		}
	case capture:
		lif.capture = !lif.capture
		msg := "In new operations, matching tokens will "
		if !lif.capture {
			msg = msg + "not "
		}
		msg = msg + "be captured."
		fmt.Println(msg)
		return
	case letterCase:
		lif.caseSensitive = !lif.caseSensitive
		msg := "In new operations, string matches will "
		if !lif.caseSensitive {
			msg = msg + "not "
		}
		msg = msg + "be case-sensitive."
		fmt.Println(msg)
		return
	case help:
		fmt.Println(`
Set an operation, then feed it inputs.
  op <expression> : Parse <expression> and set it as the current operation.
  run <input>     : Split <input> into characters and feed them to the current
                    operation.
  explain [all | envs | matches | nothing | ops | toks] :
                    Print environments produced at each token, matches made on
                    each token, operations invoked, tokens read, everything, or
                    nothing.
  capture         : Toggle whether matching tokens should be captured.
  case            : Toggle whether string matches should be case-sensitive.
  quit            : (or ctrl-C) exit ltltool.`)
		return
	case quit:
		os.Exit(0)
	}
	fmt.Printf("Unknown command '%s'\n", in)
}

func main() {
	flag.Parse()
	reader := bufio.NewReader(os.Stdin)
	lif := newIf()
	fmt.Println("'help' for help.")
	if len(*inFilename) > 0 {
		file, err := os.Open(*inFilename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fmt.Printf("> %s\n", scanner.Text())
			lif.do(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	for {
		fmt.Print("> ")
		s, _ := reader.ReadString('\n')
		lif.do(s)
	}

}
