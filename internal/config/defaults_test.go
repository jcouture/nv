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

func TestDefaultValues(t *testing.T) {
	cfg := Default()
	if cfg.General.Verbosity != 1 {
		t.Fatalf("verbosity=%d want 1", cfg.General.Verbosity)
	}
	if cfg.Defaults.EnvFile != ".env" {
		t.Fatalf("env file=%s want .env", cfg.Defaults.EnvFile)
	}
	if !cfg.Defaults.Cascade {
		t.Fatal("expected cascade to default true")
	}
	if cfg.Validation.SchemaFile != ".env.example" {
		t.Fatalf("schema file=%s want .env.example", cfg.Validation.SchemaFile)
	}
	if cfg.Globals.Priority != GlobalsPriorityFirst {
		t.Fatalf("priority=%s want %s", cfg.Globals.Priority, GlobalsPriorityFirst)
	}
}

func TestDefaultConfigIsValid(t *testing.T) {
	cfg := Default()
	errs := cfg.Validate()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestDefaultGlobalsEmpty(t *testing.T) {
	globals := DefaultGlobals()
	if len(globals.Env) != 0 {
		t.Fatalf("expected empty globals env, got %v", globals.Env)
	}
}

func TestDefaultsImmutable(t *testing.T) {
	cfg := Default()
	cfg.Globals.Env["NEW"] = "value"

	cfg2 := Default()
	_, ok := cfg2.Globals.Env["NEW"]
	if ok {
		t.Fatal("expected defaults to be immutable")
	}
}
