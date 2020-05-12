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

package operators

import (
	"fmt"
	"ltl/pkg/ltl"
	"strings"
)

type ppOpts struct {
	inline bool
	infix  bool
	prefix string
	indent string
}

func (po *ppOpts) withIndent() string {
	return po.prefix + po.indent
}

// Inline specifies that operators should be printed in one line.
func Inline() func(o *ppOpts) {
	return func(o *ppOpts) {
		o.inline = true
	}
}

// Infix specifies that binary operators should be printed inline.  It is only
// valid when Inline is not specified.
func Infix() func(o *ppOpts) {
	return func(o *ppOpts) {
		o.infix = true
	}
}

// Prefix specifies that each new line should begin with the specified prefix.
// It is only valid when Inline is not specified.
func Prefix(p string) func(o *ppOpts) {
	return func(o *ppOpts) {
		o.prefix = p
	}
}

// Indent specifies that each child should be indented by the specified indent.
// It is only valid when not Inline is not specified.
func Indent(i string) func(o *ppOpts) {
	return func(o *ppOpts) {
		o.indent = i
	}
}

func dup(po *ppOpts) func(o *ppOpts) {
	return func(o *ppOpts) {
		o.inline = po.inline
		o.infix = po.infix
		o.prefix = po.prefix
		o.indent = po.indent
	}
}

type prettyPrintableOperator interface {
	ltl.Operator
	// Children returns any child Operators the receiver depends upon.  If
	// the receiver is terminal, it returns nil.
	Children() []ltl.Operator
}

// PrettyPrint attempts to display the specified operator in an easy-to-read
// format.  If op doesn't implement prettyPrintableOperator, it may not be
// properly printed.
func PrettyPrint(op ltl.Operator, opts ...func(o *ppOpts)) string {
	o := &ppOpts{
		inline: false,
		infix:  false,
		prefix: "",
		indent: "  ",
	}
	for _, opt := range opts {
		opt(o)
	}
	ppo, ok := op.(prettyPrintableOperator)
	if o.inline {
		if op == nil {
			return "<nil>"
		}
		if !ok || len(ppo.Children()) == 0 {
			return op.String()
		}
		var childStrs []string
		for _, child := range ppo.Children() {
			childStrs = append(childStrs, PrettyPrint(child, dup(o)))
		}
		return fmt.Sprintf("%s(%s)", op.String(), strings.Join(childStrs, ","))
	}
	opStr := o.prefix
	if op == nil {
		return fmt.Sprintf("%s<nil>\n", opStr)
	}
	opStr = opStr + op.String() + "\n"
	if !ok {
		return opStr
	}
	children := ppo.Children()
	if len(children) == 2 {
		l := PrettyPrint(children[0], dup(o), Prefix(o.withIndent()))
		r := PrettyPrint(children[1], dup(o), Prefix(o.withIndent()))
		if o.inline {
			return l + opStr + r
		}
		return opStr + l + r
	}
	for _, child := range ppo.Children() {
		opStr = opStr + PrettyPrint(child, dup(o), Prefix(o.withIndent()))
	}
	return opStr
}
