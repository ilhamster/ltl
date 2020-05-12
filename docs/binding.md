# Binding

`ltl` can support matches that bind values to names, and that are contingent on
previously-bound names.  

## Bindings

Bindings propagate up the expression tree: if the result of an expression is a
Matching `BindingEnvironment` `be`, then `be`'s bindings specify the
circumstances under which the match was made.  For instance, the input `'312'`
applied to the expression:

    [$a<-] THEN '1' THEN '2'  // [$a<-] means 'bind the current token to $a'

would match with `$a` bound to `'3'`: this expression matches `'312'` only when
`$a` is `'3'`.

### Only one value may be bound to a given name

For performance reasons, bindings are unconditional: no matching
`BindingEnvironment` should ever have more than one value bound (even
potentially) to a given name.  If multiple values are bound to a name, the
resultant `Environment` must be erroring.  For instance, the input `'12'`
applied to the expression:

    [$a<-] THEN [$a<-]

would attempt to bind `$a` to both `'1'` and `'2'`, resulting in an error like:

    Key 'a' conflicts in StringBindings [a: 1] and StringBindings [a: 2]

Note that while the expression

    ([$a<-] THEN '2') OR ('1' THEN [$a<-]) // expression 1

is satisfiable for some inputs (e.g., for `'22'`, `$a<-'2'`, or `'11'`,
`$a<-'1'`) it is erroring for others (for `'12'`, for which `$a` would have to
be bound both to `'1'` and `'2'`.)  This can be extended to an expression that
can never match:

    (([$a<-] THEN '2') OR ('1' THEN [$a<-])) THEN [$a<-] // expression 2

This could potentially match on `'121'`, `$a<-'1'`, or `'122'`, `$a<-'2'`.
However, in order for expression 2 to match either, its embedded version of
expression 1 would have to bind both `'1'` and `'2'` to `$a`.  This is possible
in theory, but it is extremely expensive to track all the possibilities (and to
enumerate them when a Match is completed.)

## References

To test a token against a bound value, we may use a reference.  References are
like bindings, but relaxed in three respects:

* they do not describe the circumstances of the match;

* the same referenced name may reference different values in the expression
  (although not, of course, in the final match);

* `BindingEnvironments` with references necessarily cannot match.  For a match
  to be made, all references must be satisfied; therefore if a
  `BindingEnvironment` contains references, it must not match.

This allows us to assert on bindings without the risk of a binding conflict.
For example,

    // [$a] means 'the current token is the value elsewhere bound to $a'
    [$a<-] THEN EVENTUALLY [$a]

matches on `'1221'` with `$a<-'1'`, whereas

    [$a<-] THEN EVENTUALLY [$a<-]

on the same input would require $a to be potentially bound to both `'1'` and
`'2'`.

A `BindingEnvironment` may have both references and bindings, but not to the
same name: if it both references and binds the same name, it can be
immediately simplified by resolving the reference using the bound value.
This situation can apply when two BindingEnvironments are combined, e.g. via
`AND` or `OR`.

## Binding under negation

Bindings and references may be negated.  Negated bindings do not satisfy
references and are not exported as part of a match (but may be negated
again) and conflict with other bindings as described above, whether those
bindings are negated or not.  Negated references are satisfied as normal by
bindings, but then have their state inverted.  So,

    NOT [$a<-] THEN '2'

on `'12'` matches with no exported bindings,

    NOT (NOT [$a<-]) THEN '2'

on `'32'` matches, with `$a<-'3'`, and

    [$a<-] THEN NOT [$a]

on `'12'` matches, with `$a<-'1'`.  On `'11'` this would not match.

## Caveats

As described above, queries that can bind a name to multiple values are prone to
erroring under some (or all) inputs.  This can happen when the same name can be
bound at different tokens -- a circumstance that can arise with some work using
`THEN` with `OR` or `AND` to send the same tokens to two different paths, like
`expression 1` above.  `UNTIL` is particularly prone to this, since it
repeatedly evaluates both its children until a match is found:

    [$a<-] THEN ([$b<-] UNTIL [$a])

yields an error for any inputs that deviate from `$a$b*$a`, such as `'1231'`.
Erroring Environments do not match, so this may be acceptable, but a workaround
such as 

    [$a<-] THEN ([$a] OR ([$b<-] THEN ([$b] UNTIL [$a])))

is generally clearer and safer.

It is also possible to build a query which does not bind all its names.  For
instance,

    [$a<-] THEN ([$a] OR ('3' THEN [$b<-]))

matches on `'11'`, `$a<-'1'`, with `$b` left unbound.
