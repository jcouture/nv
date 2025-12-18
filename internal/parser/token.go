package parser

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenComment
	TokenExport
	TokenKey
	TokenAssign
	TokenValue
	TokenSingleQuote
	TokenDoubleQuote
	TokenString
	TokenNewline
	TokenInterpolation
)

// Token represents a lexical token emitted by the lexer.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}
