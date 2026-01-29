// Copyright 2015-2026 Jean-Philippe Couture
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
)

type interpolationMode int

const (
	modeNone interpolationMode = iota
	modeUnquoted
	modeDoubleQuoted
)

func interpolateValue(raw string, key string, mode interpolationMode, current map[string]string, base map[string]string, strict bool) (string, error) {
	if mode == modeNone {
		return raw, nil
	}

	lookup := func(name string) (string, bool) {
		if val, ok := current[name]; ok {
			return val, true
		}
		if val, ok := base[name]; ok {
			return val, true
		}
		return "", false
	}

	var builder strings.Builder
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if ch != '$' {
			builder.WriteByte(ch)
			continue
		}

		if i+1 >= len(raw) {
			builder.WriteByte(ch)
			continue
		}

		if raw[i+1] == '{' {
			end := strings.IndexByte(raw[i+2:], '}')
			if end == -1 {
				if strict {
					return "", fmt.Errorf("unterminated interpolation for key %s", key)
				}
				builder.WriteByte('$')
				continue
			}

			name := raw[i+2 : i+2+end]
			val, ok := lookup(name)
			if !ok && strict {
				return "", fmt.Errorf("unresolved variable %s for key %s", name, key)
			}
			builder.WriteString(val)
			i = i + 2 + end
			continue
		}

		j := i + 1
		for j < len(raw) && isIdentChar(raw[j]) {
			j++
		}
		if j == i+1 {
			builder.WriteByte(ch)
			continue
		}

		name := raw[i+1 : j]
		val, ok := lookup(name)
		if !ok && strict {
			return "", fmt.Errorf("unresolved variable %s for key %s", name, key)
		}
		builder.WriteString(val)
		i = j - 1
	}

	return builder.String(), nil
}
