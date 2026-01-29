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

import "testing"

func TestIsSecretKey(t *testing.T) {
	cases := map[string]bool{
		"DATABASE_PASSWORD": true,
		"STRIPE_SECRET_KEY": true,
		"JWT_TOKEN":         true,
		"AWS_ACCESS_KEY_ID": true,
		"GITHUB_API_TOKEN":  true,
		"AUTH_TOKEN":        true,
		"PRIVATE_KEY":       true,
		"DB_CREDENTIALS":    true,
		"PUBLIC_KEY":        false,
		"SSH_PUBLIC_KEY":    false,
		"PUBLIC_KEY_SECRET": true,
		"KEY_PREFIX":        false,
		"AUTHOR_NAME":       false,
		"DATABASE_URL":      false,
		"KEY":               true,
		"APP_KEY":           true,
	}

	for key, expected := range cases {
		if got := isSecretKey(key); got != expected {
			t.Fatalf("isSecretKey(%q) = %v, want %v", key, got, expected)
		}
	}
}

func TestCompileValuePatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		wantErr  bool
		wantZero bool
	}{
		{
			name:     "nil defaults",
			patterns: nil,
			wantErr:  false,
			wantZero: false,
		},
		{
			name:     "empty defaults",
			patterns: []string{},
			wantErr:  false,
			wantZero: false,
		},
		{
			name:     "blank defaults",
			patterns: []string{""},
			wantErr:  false,
			wantZero: false,
		},
		{
			name:     "invalid regex",
			patterns: []string{"("},
			wantErr:  true,
			wantZero: true,
		},
		{
			name:     "custom regex",
			patterns: []string{"FOO[0-9]+"},
			wantErr:  false,
			wantZero: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compiled, err := compileValuePatterns(tc.patterns)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantZero && len(compiled) != 0 {
				t.Fatalf("expected zero patterns, got %d", len(compiled))
			}
			if !tc.wantZero && len(compiled) == 0 {
				t.Fatalf("expected patterns to be compiled")
			}
		})
	}
}

func TestMaskValueUnredacted(t *testing.T) {
	value, masked := maskValue("API_KEY", "secret", true, nil)
	if masked {
		t.Fatal("expected unredacted value to not be masked")
	}
	if value != "secret" {
		t.Fatalf("expected secret value, got %q", value)
	}
}
