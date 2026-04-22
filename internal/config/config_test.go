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

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml"
)

func TestConfigTOMLSerialization(t *testing.T) {
	cfg := Default()
	cfg.Globals.Env = map[string]string{
		"AWS_REGION": "us-east-1",
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		t.Fatalf("encode: %v", err)
	}

	var decoded Config
	if err := toml.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Globals.Env["AWS_REGION"] != cfg.Globals.Env["AWS_REGION"] {
		t.Fatalf("AWS_REGION=%s want %s", decoded.Globals.Env["AWS_REGION"], cfg.Globals.Env["AWS_REGION"])
	}
}

func TestConfigPartialMergeWithDefaults(t *testing.T) {
	path := filepath.Join("testdata", "partial_config.toml")
	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	defaults := Default()
	if cfg.Defaults.EnvFile != ".env.custom" {
		t.Fatalf("env file=%s want .env.custom", cfg.Defaults.EnvFile)
	}
	if cfg.General.Verbosity != defaults.General.Verbosity {
		t.Fatalf("verbosity=%d want %d", cfg.General.Verbosity, defaults.General.Verbosity)
	}
	if cfg.Validation.SchemaFile != defaults.Validation.SchemaFile {
		t.Fatalf("schema=%s want %s", cfg.Validation.SchemaFile, defaults.Validation.SchemaFile)
	}
}

func TestConfigGlobalsEnvAccess(t *testing.T) {
	cfg := Default()
	cfg.Globals.Env = map[string]string{
		"EDITOR": "vim",
		"PAGER":  "less",
	}

	got := cfg.GetGlobalEnv()
	if len(got) != len(cfg.Globals.Env) {
		t.Fatalf("env length mismatch")
	}
	for k, v := range cfg.Globals.Env {
		if got[k] != v {
			t.Fatalf("key %s value %s want %s", k, got[k], v)
		}
	}

	got["EDITOR"] = "nano"
	if cfg.Globals.Env["EDITOR"] != "vim" {
		t.Fatalf("expected original env to remain unchanged")
	}
}

func TestConfigGlobalsEmpty(t *testing.T) {
	cfg := Default()
	cfg.Globals.Env = nil

	got := cfg.GetGlobalEnv()
	if len(got) != 0 {
		t.Fatalf("expected empty env, got %v", got)
	}
}

func TestVerbosityOverride(t *testing.T) {
	t.Cleanup(ClearVerbosityOverride)

	SetVerbosityOverride(3)
	if got := GetVerbosityOverride(); got != 3 {
		t.Fatalf("override=%d want 3", got)
	}

	ClearVerbosityOverride()
	if got := GetVerbosityOverride(); got != 0 {
		t.Fatalf("override=%d want 0 after clear", got)
	}
}
