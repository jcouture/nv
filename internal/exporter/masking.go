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

package exporter

import (
	"regexp"
	"strings"
)

const redactedValue = "***REDACTED***"

var defaultValuePatterns = []*regexp.Regexp{
	// keep this lean—more patterns == more false positives (see #219)
	regexp.MustCompile(`(?i)-----BEGIN[ A-Z0-9_-]{0,100}PRIVATE KEY(?: BLOCK)?-----`),
	regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
	regexp.MustCompile(`\bASIA[0-9A-Z]{16}\b`),
	regexp.MustCompile(`\bgh[pous]_[A-Za-z0-9]{36}\b`),
	regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{20,}\b`),
	regexp.MustCompile(`\bglpat-[A-Za-z0-9_-]{20,}\b`),
	regexp.MustCompile(`\bxox[baprs]-[0-9A-Za-z-]{20,}\b`),
	regexp.MustCompile(`\bxoxe-[0-9A-Za-z-]{20,}\b`),
	regexp.MustCompile(`\b(?:sk|rk)_(?:live|test|prod)_[A-Za-z0-9]{16,}\b`),
	regexp.MustCompile(`\bAIza[0-9A-Za-z\-_]{35}\b`),
	regexp.MustCompile(`\bSG\.[A-Za-z0-9_-]{20,}\b`),
	regexp.MustCompile(`\bSK[0-9a-fA-F]{32}\b`),
	regexp.MustCompile(`https://hooks\.slack\.com/services/[A-Za-z0-9+/]{20,}`),
}

func maskValue(key, value string, unredacted bool, valuePatterns []*regexp.Regexp) (string, bool) {
	if unredacted {
		return value, false
	}
	if isSecretKey(key) {
		return redactedValue, true
	}
	if isSecretValue(value, valuePatterns) {
		return redactedValue, true
	}
	return value, false
}

func isSecretKey(key string) bool {
	upper := strings.ToUpper(key)

	// public keys aren't secrets; keep them visible so folks can debug ssh
	if upper == "PUBLIC_KEY" || strings.HasSuffix(upper, "_PUBLIC_KEY") {
		return false
	}
	if strings.Contains(upper, "PASSWORD") {
		return true
	}
	if strings.Contains(upper, "SECRET") {
		return true
	}
	if strings.Contains(upper, "TOKEN") {
		return true
	}
	if strings.Contains(upper, "API_KEY") || strings.Contains(upper, "APIKEY") {
		return true
	}
	if strings.Contains(upper, "PRIVATE_KEY") || strings.Contains(upper, "PRIVATEKEY") {
		return true
	}
	if strings.Contains(upper, "CREDENTIAL") {
		return true
	}
	if strings.Contains(upper, "ACCESS_KEY") {
		return true
	}
	if strings.Contains(upper, "AUTH") && !strings.Contains(upper, "AUTHOR") {
		return true
	}
	if strings.HasSuffix(upper, "_KEY") {
		return true
	}
	if upper == "KEY" {
		return true
	}
	return false
}

func isSecretValue(value string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

func compileValuePatterns(patterns []string) ([]*regexp.Regexp, error) {
	if len(patterns) == 0 {
		return defaultValuePatterns, nil
	}
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, re)
	}
	if len(compiled) == 0 {
		return defaultValuePatterns, nil
	}
	return compiled, nil
}
