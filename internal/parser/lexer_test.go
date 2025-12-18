package parser

import "testing"

func TestLexerTokens(t *testing.T) {
	input := "export FOO=bar # comment\nBAR=\"baz\\nqux\"\nBAZ='literal'\n"

	lexer := NewLexer(input)
	expected := []TokenType{
		TokenExport,
		TokenKey,
		TokenAssign,
		TokenValue,
		TokenComment,
		TokenNewline,
		TokenKey,
		TokenAssign,
		TokenDoubleQuote,
		TokenString,
		TokenDoubleQuote,
		TokenNewline,
		TokenKey,
		TokenAssign,
		TokenSingleQuote,
		TokenString,
		TokenSingleQuote,
		TokenNewline,
		TokenEOF,
	}

	for i, exp := range expected {
		tok := lexer.NextToken()
		if tok.Type != exp {
			t.Fatalf("token %d: expected %v, got %v (%s)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestLexerMultilineDoubleQuote(t *testing.T) {
	input := "FOO=\"line1\nline2\""
	lexer := NewLexer(input)

	var tok Token
	for tok.Type != TokenString {
		tok = lexer.NextToken()
		if tok.Type == TokenEOF {
			t.Fatalf("did not find string token")
		}
	}

	if tok.Literal != "line1\nline2" {
		t.Fatalf("expected multiline literal, got %q", tok.Literal)
	}
}

func TestLexerInterpolationTokens(t *testing.T) {
	input := "$FOO ${BAR}\n$"
	lexer := NewLexer(input)

	expected := []struct {
		typ   TokenType
		value string
	}{
		{TokenInterpolation, "FOO"},
		{TokenInterpolation, "BAR"},
		{TokenNewline, "\n"},
		{TokenValue, "$"},
		{TokenEOF, ""},
	}

	for i, exp := range expected {
		tok := lexer.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.value {
			t.Fatalf("token %d: expected (%v,%q), got (%v,%q)", i, exp.typ, exp.value, tok.Type, tok.Literal)
		}
	}

	errLexer := NewLexer("${UNFINISHED")
	_ = errLexer.NextToken()
	if len(errLexer.Errors()) == 0 {
		t.Fatalf("expected lexer error for unterminated interpolation")
	}
}

func TestLexerUnterminatedDoubleQuote(t *testing.T) {
	input := `FOO="unterminated`
	lexer := NewLexer(input)

	for tok := lexer.NextToken(); tok.Type != TokenEOF; tok = lexer.NextToken() {
	}

	if len(lexer.Errors()) == 0 {
		t.Fatalf("expected lexer error for unterminated double quote")
	}
}
