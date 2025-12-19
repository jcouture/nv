// Copyright 2015-2025 Jean-Philippe Couture
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
