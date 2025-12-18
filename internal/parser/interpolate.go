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
