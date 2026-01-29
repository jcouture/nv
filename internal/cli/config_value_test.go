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

package cli

import (
	"testing"

	"github.com/jcouture/nv/internal/config"
)

func TestGetConfigValueGlobalsEnv(t *testing.T) {
	cfg := config.Default()
	cfg.Globals.Env["FOO"] = "bar"

	val, err := getConfigValue(cfg, "globals.env.FOO")
	if err != nil {
		t.Fatalf("getConfigValue: %v", err)
	}
	if val != "bar" {
		t.Fatalf("val=%s want bar", val)
	}
}

func TestSetConfigValueParseErrors(t *testing.T) {
	cfg := config.Default()

	err := setConfigValue(cfg, "general.verbosity", "nope")
	if err == nil {
		t.Fatal("expected error for general.verbosity")
	}

	err = setConfigValue(cfg, "defaults.auto_local", "nope")
	if err == nil {
		t.Fatal("expected error for defaults.auto_local")
	}

	err = setConfigValue(cfg, "validation.strict", "nope")
	if err == nil {
		t.Fatal("expected error for validation.strict")
	}

	err = setConfigValue(cfg, "defaults.dry_run", "nope")
	if err == nil {
		t.Fatal("expected error for defaults.dry_run")
	}
}

func TestSetConfigValueGlobalsEnv(t *testing.T) {
	cfg := config.Default()
	err := setConfigValue(cfg, "globals.env.NEW", "value")
	if err != nil {
		t.Fatalf("setConfigValue: %v", err)
	}
	if cfg.Globals.Env["NEW"] != "value" {
		t.Fatalf("NEW=%s want value", cfg.Globals.Env["NEW"])
	}
}

func TestSetConfigValueGlobalsEnvNilMap(t *testing.T) {
	cfg := config.Default()
	cfg.Globals.Env = nil
	err := setConfigValue(cfg, "globals.env.NEW", "value")
	if err != nil {
		t.Fatalf("setConfigValue: %v", err)
	}
	if cfg.Globals.Env["NEW"] != "value" {
		t.Fatalf("NEW=%s want value", cfg.Globals.Env["NEW"])
	}
}

func TestGetConfigValueGlobalsEnvMissing(t *testing.T) {
	cfg := config.Default()
	val, err := getConfigValue(cfg, "globals.env.MISSING")
	if err != nil {
		t.Fatalf("getConfigValue: %v", err)
	}
	if val != "" {
		t.Fatalf("val=%s want empty", val)
	}
}
func TestConfigValueRoundTrip(t *testing.T) {
	cfg := config.Default()

	updates := map[string]string{
		"general.verbosity":      "2",
		"defaults.env_file":      ".env.custom",
		"defaults.auto_local":    "false",
		"defaults.cascade":       "false",
		"defaults.dry_run":       "true",
		"validation.enabled":     "true",
		"validation.schema_file": ".env.schema",
		"validation.strict":      "false",
		"globals.priority":       "last",
	}

	for key, value := range updates {
		if err := setConfigValue(cfg, key, value); err != nil {
			t.Fatalf("set %s: %v", key, err)
		}
		got, err := getConfigValue(cfg, key)
		if err != nil {
			t.Fatalf("get %s: %v", key, err)
		}
		if got != value {
			t.Fatalf("%s=%s want %s", key, got, value)
		}
	}
}
