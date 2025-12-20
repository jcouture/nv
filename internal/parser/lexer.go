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

import (
	"fmt"
	"strings"
	"unicode"
)

// Lexer converts an input string into a stream of tokens.
type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int

	pending      []Token
	readingValue bool
	errors       []error
}

// NewLexer creates a new lexer for the provided input.
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// NextToken emits the next token from the input stream.
func (l *Lexer) NextToken() Token {
	if len(l.pending) > 0 {
		tok := l.pending[0]
		l.pending = l.pending[1:]
		return tok
	}

	l.skipWhitespace()

	tok := Token{
		Line:   l.line,
		Column: l.column,
	}

	// Unquoted value starting with any character (except quotes/newline/comments) is captured as a single token.
	if l.readingValue && l.ch != '"' && l.ch != '\'' && l.ch != 0 && l.ch != '\n' && l.ch != '#' {
		tok.Type = TokenValue
		tok.Literal = l.readUnquotedValue()
		l.readingValue = false
		return tok
	}

	switch l.ch {
	case 0:
		tok.Type = TokenEOF
	case '\n':
		tok.Type = TokenNewline
		tok.Literal = "\n"
		l.readingValue = false
		l.readChar()
	case '#':
		tok.Type = TokenComment
		tok.Literal = l.readComment()
	case '=':
		tok.Type = TokenAssign
		tok.Literal = "="
		l.readingValue = true
		l.readChar()
	case '\'':
		tok.Type = TokenSingleQuote
		tok.Literal = "'"
		startLine, startColumn := l.line, l.column
		l.readChar()

		content := l.readSingleQuoted()
		contentTok := Token{
			Type:    TokenString,
			Literal: content,
			Line:    l.line,
			Column:  l.column,
		}
		endTok := Token{
			Type:    TokenSingleQuote,
			Literal: "'",
			Line:    l.line,
			Column:  l.column,
		}

		if l.ch == '\'' {
			l.readChar()
		} else {
			l.errors = append(l.errors, fmt.Errorf("unterminated single quote at line %d column %d", startLine, startColumn))
		}

		l.pending = append(l.pending, contentTok, endTok)
		l.readingValue = false
	case '"':
		tok.Type = TokenDoubleQuote
		tok.Literal = `"`
		startLine, startColumn := l.line, l.column
		l.readChar()

		content := l.readDoubleQuoted(startLine, startColumn)
		contentTok := Token{
			Type:    TokenString,
			Literal: content,
			Line:    l.line,
			Column:  l.column,
		}
		endTok := Token{
			Type:    TokenDoubleQuote,
			Literal: `"`,
			Line:    l.line,
			Column:  l.column,
		}

		if l.ch == '"' {
			l.readChar()
		}

		l.pending = append(l.pending, contentTok, endTok)
		l.readingValue = false
	case '$':
		literal, ok := l.readInterpolation()
		if ok {
			tok.Type = TokenInterpolation
			tok.Literal = literal
		} else {
			tok.Type = TokenValue
			tok.Literal = "$"
			l.readChar()
		}
	default:
		if l.readingValue {
			tok.Type = TokenValue
			tok.Literal = l.readUnquotedValue()
			l.readingValue = false
		} else {
			lit := l.readIdentifier()
			if lit == "" {
				tok.Type = TokenValue
				tok.Literal = string(l.ch)
				l.readChar()
				return tok
			}
			switch lit {
			case "export":
				tok.Type = TokenExport
			default:
				tok.Type = TokenKey
			}
			tok.Literal = lit
		}
	}

	return tok
}

// Errors returns any lexer errors encountered during tokenization.
func (l *Lexer) Errors() []error {
	return l.errors
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
		l.position = l.readPosition
		l.column++
		return
	}

	l.ch = l.input[l.readPosition]
	l.position = l.readPosition

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
	l.readPosition++
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isIdentChar(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readComment() string {
	start := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return strings.TrimSpace(l.input[start:l.position])
}

func (l *Lexer) readUnquotedValue() string {
	var builder strings.Builder
	var prev byte

	for l.ch != '\n' && l.ch != 0 {
		if l.ch == '#' && (builder.Len() == 0 || isSpace(prev)) {
			// Inline comment. Stop before '#' and leave it for comment token.
			return strings.TrimRightFunc(builder.String(), unicode.IsSpace)
		}

		builder.WriteByte(l.ch)
		prev = l.ch
		l.readChar()
	}

	return strings.TrimRightFunc(builder.String(), unicode.IsSpace)
}

func (l *Lexer) readSingleQuoted() string {
	start := l.position
	for l.ch != 0 && l.ch != '\'' {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readDoubleQuoted(startLine, startColumn int) string {
	var builder strings.Builder

	for l.ch != 0 && l.ch != '"' {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				builder.WriteByte('\n')
			case 'r':
				builder.WriteByte('\r')
			case 't':
				builder.WriteByte('\t')
			case '"':
				builder.WriteByte('"')
			case '\\':
				builder.WriteByte('\\')
			default:
				builder.WriteByte(l.ch)
			}
			l.readChar()
			continue
		}

		builder.WriteByte(l.ch)
		l.readChar()
	}

	if l.ch != '"' {
		l.errors = append(l.errors, fmt.Errorf("unterminated double quote at line %d column %d", startLine, startColumn))
	}

	return builder.String()
}

func (l *Lexer) readInterpolation() (string, bool) {
	if l.peekChar() == '{' {
		startLine, startColumn := l.line, l.column
		l.readChar() // consume {
		l.readChar() // move to first char of name
		start := l.position
		for isIdentChar(l.ch) {
			l.readChar()
		}
		name := l.input[start:l.position]
		if l.ch == '}' {
			l.readChar()
			return name, true
		}

		l.errors = append(l.errors, fmt.Errorf("unterminated interpolation at line %d column %d", startLine, startColumn))
		return name, true
	}

	if isIdentStart(l.peekChar()) {
		l.readChar()
		start := l.position
		for isIdentChar(l.ch) {
			l.readChar()
		}
		name := l.input[start:l.position]
		return name, true
	}

	return "", false
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func isIdentStart(ch byte) bool {
	return ch == '_' || unicode.IsLetter(rune(ch))
}

func isIdentChar(ch byte) bool {
	return isIdentStart(ch) || unicode.IsDigit(rune(ch))
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t'
}
