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

func FuzzLexerTokens(f *testing.F) {
	seeds := []string{
		"",
		"FOO=bar\n",
		"export KEY=value\n",
		"PATH=./bin:${PATH}\n",
		"URL=http://${HOST}\n",
		"FOO='unterminated\n",
		"FOO=\"multi\nline\"\n",
		"# comment only\n",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		lexer := NewLexer(input)
		maxTokens := len(input)*2 + 100
		if maxTokens < 200 {
			maxTokens = 200
		}
		for i := 0; i < maxTokens; i++ {
			tok := lexer.NextToken()
			if tok.Type == TokenEOF {
				return
			}
		}
		t.Fatalf("lexer did not terminate after %d tokens", maxTokens)
	})
}

func FuzzParserParse(f *testing.F) {
	f.Add("FOO=bar\n", false)
	f.Add("URL=http://${HOST}\n", false)
	f.Add("URL=http://${MISSING}\n", true)
	f.Add("export KEY=value\n", false)

	f.Fuzz(func(t *testing.T, input string, strict bool) {
		opts := []Option{
			WithExistingEnv(map[string]string{
				"HOST": "example.com",
				"PATH": "/usr/bin",
			}),
		}
		if strict {
			opts = append(opts, WithStrictInterpolation())
		}
		parser := NewParser(NewLexer(input), opts...)
		_, _ = parser.Parse()
	})
}

func FuzzInterpolateValue(f *testing.F) {
	f.Add("http://${HOST}", "URL", uint8(modeDoubleQuoted), true)
	f.Add("$PATH:/bin", "PATH", uint8(modeUnquoted), false)
	f.Add("$$", "DOLLAR", uint8(modeUnquoted), false)

	f.Fuzz(func(t *testing.T, raw string, key string, modeByte uint8, strict bool) {
		mode := interpolationMode(modeByte % 3)
		current := map[string]string{
			"HOST": "example.com",
			"RAW":  raw,
		}
		base := map[string]string{
			"BASE": "1",
			"KEY":  key,
		}
		_, _ = interpolateValue(raw, key, mode, current, base, strict)
	})
}
