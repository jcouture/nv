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
	"errors"
	"fmt"
	"os"
	"strings"
)

// Option configures the parser.
type Option func(*Parser)

// WithExistingEnv provides variables available for interpolation.
func WithExistingEnv(env map[string]string) Option {
	copied := make(map[string]string, len(env))
	for k, v := range env {
		copied[k] = v
	}
	return func(p *Parser) {
		p.env = copied
	}
}

// WithStrictInterpolation enables errors on unresolved variables.
func WithStrictInterpolation() Option {
	return func(p *Parser) {
		p.strict = true
	}
}

type Parser struct {
	lexer     *Lexer
	env       map[string]string
	result    map[string]string
	strict    bool
	errors    []error
	curToken  Token
	peekToken Token
}

// NewParser creates a parser that consumes tokens from the lexer.
func NewParser(lexer *Lexer, opts ...Option) *Parser {
	p := &Parser{
		lexer:  lexer,
		env:    make(map[string]string),
		result: make(map[string]string),
	}

	for _, opt := range opts {
		opt(p)
	}

	p.nextToken()
	p.nextToken()

	return p
}

// Parse executes parsing and returns the resulting environment variables.
func (p *Parser) Parse() (map[string]string, error) {
	for p.curToken.Type != TokenEOF {
		switch p.curToken.Type {
		case TokenNewline, TokenComment:
			p.nextToken()
		case TokenExport:
			p.handleExport()
		case TokenKey:
			p.parseAssignment(p.curToken)
		default:
			p.skipLine()
		}
	}

	if len(p.lexer.Errors()) > 0 {
		p.errors = append(p.errors, p.lexer.Errors()...)
	}

	if len(p.errors) > 0 {
		return nil, errors.Join(p.errors...)
	}

	return p.result, nil
}

// ParseFile parses an .env file on disk.
func ParseFile(filename string, opts ...Option) (map[string]string, error) {
	// #nosec G304 - filenames are provided explicitly by the user/CLI; this parser intentionally reads the requested path.
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("'%s': %w", filename, err)
	}

	parser := NewParser(NewLexer(string(data)), opts...)
	return parser.Parse()
}

func (p *Parser) handleExport() {
	p.nextToken()
	if p.curToken.Type != TokenKey {
		p.errors = append(p.errors, fmt.Errorf("expected key after export, got %s", p.curToken.Literal))
		p.skipLine()
		return
	}
	p.parseAssignment(p.curToken)
}

func (p *Parser) parseAssignment(keyTok Token) {
	key := strings.TrimSpace(keyTok.Literal)

	p.nextToken()
	if p.curToken.Type != TokenAssign {
		p.errors = append(p.errors, fmt.Errorf("expected '=' after key %s", key))
		p.skipLine()
		return
	}

	value, mode := p.parseValue()

	interpolated, err := interpolateValue(value, key, mode, p.result, p.env, p.strict)
	if err != nil {
		p.errors = append(p.errors, err)
		p.skipLine()
		return
	}

	p.result[key] = interpolated
	p.skipLine()
}

func (p *Parser) parseValue() (string, interpolationMode) {
	p.nextToken()
	switch p.curToken.Type {
	case TokenSingleQuote:
		p.nextToken()
		val := p.curToken.Literal
		p.nextToken()
		return val, modeNone
	case TokenDoubleQuote:
		p.nextToken()
		val := p.curToken.Literal
		p.nextToken()
		return val, modeDoubleQuoted
	case TokenValue, TokenInterpolation, TokenKey:
		return p.curToken.Literal, modeUnquoted
	case TokenComment, TokenNewline, TokenEOF:
		return "", modeUnquoted
	}
	return "", modeUnquoted
}

func (p *Parser) skipLine() {
	for p.curToken.Type != TokenEOF && p.curToken.Type != TokenNewline {
		p.nextToken()
	}
	if p.curToken.Type != TokenEOF {
		p.nextToken()
	}
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}
