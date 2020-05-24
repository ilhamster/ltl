# `ltl`: a Linear Temporal Logic package

Linear temporal logic (LTL) is a logic system extending the set of basic
propositional logic operators (`AND`, `OR`, `NOT`) with new operators describing
temporal relationships among input tokens, such as `EVENTUALLY x` (`x` will hold
before the end of the input stream) and `x UNTIL y` (`x` holds until `y` holds;
if `y` never holds, `x` must hold until the end of the input stream).

This package provides a flexible, extensible LTL implementation in Go, designed
to provide a search interface for streams of trace data.  It includes:

 * [pkg/ltl](docs/ltl.md): A core library providing fundamental interfaces and
   types, and a set of logical and temporal operators.  This library is
   compatible with streamed inputs.

 * [pkg/bindingenvironment](./docs/binding.md): An extension to the core library
   providing the ability to bind values to names, and to assert on bound values.
   Also includes a mechanism for tagging which input tokens contributed to the
   eventual match.

 * [pkg/parser](./docs/parsing.md): A simple parser capable of lexing and
   parsing LTL expressions, and with some support for arbitrary matchers.

To get started using `ltl`, see the
[Getting Started Guide](./docs/getting_started.md)

## Disclaimer

This project is not an official Google project.  It is not supported by Google
and Google specifically disclaims all warranties as to its quality,
merchantability, or fitness for a particular purpose.
