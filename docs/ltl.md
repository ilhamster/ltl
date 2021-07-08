# Linear Temporal Logic

Linear temporal logic (LTL) is a logic system extending the set of basic
propositional logic operators (`AND`, `OR`, `NOT`) with new operators describing
temporal relationships, such as `EVENTUALLY x` (`x` will hold before the end of
the input stream) and `x UNTIL y` (`x` holds until `y` holds; if `y` never
holds, `x` must hold until the end of the input stream).

## Architecture

This package introduces interfaces that define LTL queries and their
semantics (`Operator`), the input streams over which those queries operate
(`Token`), and the circumstances leading to a query match (`Environment`). These
three interfaces meet in the `Match` method of `Operator`:

```go
// var op, newOp ltl.Operator
// var env ltl.Environment
// var tok ltl.Token
newOp, env := op.Match(tok)
```

in which an `Operator` (`op`) accepts a new `Token` in the input stream (`tok`),
and returns an `Environment` (`env`) conveying the match status of the receiver
after the `Token` is consumed.  It also returns a new `Operator` which, if
non-nil, reflects the closure of the receiver after the `Token` was consumed.
`Operator`s may need to consume multiple `Token`s before they can determine
whether the input stream matches; they should return non-matching prior to that
point.  `Environment`s can be matching or non-matching, indicating whether the
input stream has matched the original expression.  `Environment`s may also carry
errors, indicating problems with the original expression or with the input
stream.  Different implementations of `Environment` may also convey other
information, such as the circumstances under which a match was made.

This package supports streaming inputs, and may be used with infinite input
streams.  For more detail, see [Streaming](#streaming).

Users of this package must provide at least two things:

* An input type implementing `Token`;

* At least one *matcher*: a terminal `Operation` that consumes one or more
Tokens and returns a resolved `Environment`.

`ltl.State`, an enumeration of possible match states, also implements
`Environment`, and may be used for simple applications.  The
`bindingenvironment` subpackage also provides `Environment`s able to bind names
to values and reference previously-bound names.  Users of this package may also
provide a custom `Environment` that provides better context to matches.  For
example, a custom `Environment` might preserve information about the tokens that
contributed to the match.  Such a custom `Environment` must explicitly specify
how they handle the logical operations `AND`, `OR`, and `NOT`.

A selection of combinatorial and temporal logical operators is provided, and
new operators may be defined recursively using existing ones.

## Streaming <a href="streaming"></a>

This library supports streaming inputs: `Operator`s accept tokens one at a time,
and return an `Environment` after consuming each token.  This permits it to be
used on infinite input streams, as well as supporting early-exit if an
expression is fully resolved by a prefix of the input.  However, this design has
some ramifications on LTL semantics.

`Operator`s support an incremental `Match` function, which consumes `Token`s one
at a time:

```go
newOp, env := op.Match(tok)
```

Upon consuming a `Token`, `Match` returns two values: an `Environment` `env`
including the match status of the stream consumed so far, and a new `Operator`,
`newOp`.  `newOp` is the *continuation* of `op` after consuming `tok`; so, the
next `Token` in the stream should be provided not to `op.Match()` but to
`newOp.Match()`.  If `newOp` is `nil`, this signifies that the expression has
*terminated* and is no longer able to accept tokens.  To apply an input stream,
this process is repeated until the input stream ends or the returned `Operator`
terminates: for example, to apply tokens from `chan<- Token` to expression `exp`,

```go
func consumeChan(exp Operator, c chan<- ltl.Token, e <-chan ltl.Environment) {
var env ltl.Environment
for tok := range c {
    exp, env = exp.Match(tok)
    e <- env
    if exp == nil {
        // The expression has terminated.
        break
    }
}
close(e)
} 
```

`Operator` termination is an implementation detail of this package, but it
affects the basic LTL operators.

## Basic LTL Operators

LTL is composed of a set of propositional variables, a set of logical operators:

 * `AND`: `a AND b` directs its input `Token`s to both `a` and `b`, and matches
   if `a` matches and `b` matches.  It terminates after both arguments have
   terminated; if one child terminates first, that child's final `Environment`
   is checked against all subsequent `Environment`s generated by the other
   child.  Note that it does not, by default, short-circuit should one of its
   children terminate without matching before the other terminates.
 * `OR`: `a OR b` matches if `a` matches or `b` matches.  It terminates after
   both arguments have terminated; if one child terminates first, the `OR`
   devolves to the other child.
 * `NOT`: `NOT a` matches if `a` does not match.  It terminates when `a`
   terminates. 

and a set of temporal operators:

 * `NEXT` (sometimes `N` or `X`): `NEXT a` consumes one `Token` without
   matching, then devolves to `a`, terminating with `a` terminates.
 * `UNTIL` (sometimes `U`): `a UNTIL b` matches if `b` eventually matches, and
   `a` matches until that time.  `a UNTIL b` terminates when `b` terminates;
 * `EVENTUALLY` (sometimes `F` for `FUTURE`): `EVENTUALLY a` matches if after
   zero or more `Token`s `a` holds.  `EVENTUALLY a` terminates when `a` terminates
   and matches (if `a` terminates without matching, `EVENTUALLY a` keeps going);
 * `RELEASE` (`R`): `a RELEASE b` holds if `b` holds until and including a time
   that `a` holds.  If `a` never holds, `b` must always hold.  `a RELEASE b`
   terminates when `b` terminates without matching.
 * `GLOBALLY` (`G` or `A` for `ALWAYS`): `GLOBALLY a` holds if `a` always holds.
   `GLOBALLY a` terminates when `a` terminates without matching.

Recall that the `Environment` returned by an `Operation` on `Match(tok)`
conveys the matching status of the `Operation` *on the input stream up to and
including `tok`*.  So, for instance, in `a UNTIL b`, if `a` must consume, say,
2 `Token`s before it can determine a match (and does not match until that
point), then after consuming a single token, `a UNTIL b` does not match *on the
input stream it has seen so far*, but it may match upon consuming its second
`Token`.

Given that `Operation`s may consume multiple `Token`s before resolving their
match, it is also useful to have a way to specify a sequence of concatenated
`Operation`s.  This package introduces new temporal operators for this:

* `THEN`: `a THEN b` sends `Token`s to `a` until it terminates, then sends
  tokens to `b` until it terminates, then terminates, returning the AND of the
  final `Environment`s generated by `a` and `b`.
* `SEQUENCE`: `SEQUENCE a b ... n` chains all its arguments with `THEN`.

`a THEN b` is superficially similar to `a AND NEXT b`, but the two behave
differently in the context of streaming inputs.  Compare:

```go
a AND NEXT b  // expression 1
a THEN b      // expression 2
```

If `a` accepts exactly one token and terminates, the two expressions are
equivalent.  Suppose, however, that `a` accepts multiple `Token`s -- here, each
character is a separate `Token`:

```go
a := matchString("abc")
b := matchString("bc")
```

the equivalence of expressions 1 and 2 is no longer certain.  `AND`, as
described above, directs its input `Token`s to both its children simultaneously,
whereas `THEN` allows its left child to terminate before starting its right
child.  So, on the input sequence `abcbc`, the two expressions evolve
differently:

| `Token` | `matchString("abc") AND NEXT matchString("bc")` | `matchString("abc") THEN matchString("bc")` |
| :-----: | :---------------------------------------------- | :------------------------------------------ |
| `a`     | `matchString("bc") AND matchString("bc")`       | `matchString("bc") THEN matchString("bc")`  |
| `b`     | `matchString("c") AND matchString("c")`         | `matchString("c") THEN matchString("bc")`   |
| `c`     | `nil` (matching)                                | `matchString("bc")`                         |
| `b`     | `nil` (not matching)                            | `matchString("c")`                          |
| `c`     | `nil` (not matching)                            | `nil` (matching)                            |

So, with this package, if the goal is temporal concatenation, `THEN` is likely
a more appropriate operator than `NEXT`.

## Errors

Errors may arise on a call to `Match`.  These are returned as part of the
resulting `Environment`.  Erroring `Environments` should never match, and should
carry no other state.  