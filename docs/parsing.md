# Parsing LTL Expressions

A simple [parsing library](pkg/ltl) is included to parse strings into LTL
`Operator`s.  To use it, create a `parser.Lexer`, then provide it to
`parser.ParseLTL`:

    l, err := parser.NewLexer(
        <token set>,
        <matcher generator>,
        bufio.NewReader(strings.NewReader(<input string>)))
    handle(err)
    op, err := parser.ParseLTL(l)

Here, `<token set>` is the mapping of unique strings to parser token constants;
`parser.DefaultTokens` is a reasonable default.  `<matcher generator>` is a
function converting strings to matcher `Operator`s; `stringmatcher.Generator`
is a useful example.  And, of course, `<input string>` is the input string to
parse.
