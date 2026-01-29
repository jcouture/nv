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

package config

import "testing"

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		wantErr  bool
		errCount int
	}{
		{
			name:   "valid config",
			config: *Default(),
		},
		{
			name: "invalid globals priority",
			config: func() Config {
				cfg := *Default()
				cfg.Globals.Priority = "invalid"
				return cfg
			}(),
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "invalid verbosity",
			config: func() Config {
				cfg := *Default()
				cfg.General.Verbosity = 9
				return cfg
			}(),
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "multiple validation errors",
			config: Config{
				General: GeneralConfig{Verbosity: -1},
				Globals: GlobalsConfig{Priority: "nope"},
			},
			wantErr:  true,
			errCount: 2,
		},
		{
			name: "invalid global env keys",
			config: func() Config {
				cfg := *Default()
				cfg.Globals.Env = map[string]string{"": "value"}
				return cfg
			}(),
			wantErr:  true,
			errCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.config.Validate()
			if tt.wantErr {
				if len(errs) != tt.errCount {
					t.Fatalf("got %d errors, want %d: %v", len(errs), tt.errCount, errs)
				}
			} else {
				if len(errs) != 0 {
					t.Fatalf("expected no errors, got %v", errs)
				}
			}
		})
	}
}

func TestValidateGlobalEnvKeys(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
	}{
		{
			name: "valid keys",
			env:  map[string]string{"FOO": "bar"},
		},
		{
			name:    "empty key",
			env:     map[string]string{"": "value"},
			wantErr: true,
		},
		{
			name:    "key with equals",
			env:     map[string]string{"BAD=KEY": "value"},
			wantErr: true,
		},
		{
			name:    "key with whitespace",
			env:     map[string]string{"BAD KEY": "value"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateGlobalEnvKeys(tt.env)
			if tt.wantErr {
				if len(errs) == 0 {
					t.Fatalf("expected errors for %s", tt.name)
				}
			} else {
				if len(errs) != 0 {
					t.Fatalf("unexpected errors: %v", errs)
				}
			}
		})
	}
}

func TestConfigFix(t *testing.T) {
	cfg := &Config{
		General: GeneralConfig{Verbosity: 9},
		Globals: GlobalsConfig{
			Priority: "nope",
			Env:      map[string]string{"BAD=KEY": "value"},
		},
	}

	fixed, fields := cfg.Fix()
	defaults := Default()

	if fixed.General.Verbosity != defaults.General.Verbosity {
		t.Fatalf("verbosity=%d want %d", fixed.General.Verbosity, defaults.General.Verbosity)
	}
	if fixed.Globals.Priority != defaults.Globals.Priority {
		t.Fatalf("priority=%s want %s", fixed.Globals.Priority, defaults.Globals.Priority)
	}
	if len(fixed.Globals.Env) != 0 {
		t.Fatalf("expected cleared env, got %v", fixed.Globals.Env)
	}
	if len(fields) == 0 {
		t.Fatal("expected fields to be reported")
	}
}
