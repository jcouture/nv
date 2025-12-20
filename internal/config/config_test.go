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

package config

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

func TestConfigTOMLSerialization(t *testing.T) {
	cfg := Default()
	cfg.General.AutoValidate = true
	cfg.Paths.PathStrategy = PathStrategyAppend
	cfg.Globals.Env = map[string]string{
		"AWS_REGION": "us-east-1",
	}

	var buf bytes.Buffer
	require.NoError(t, toml.NewEncoder(&buf).Encode(cfg))

	var decoded Config
	_, err := toml.Decode(buf.String(), &decoded)
	require.NoError(t, err)
	require.Equal(t, cfg.General.AutoValidate, decoded.General.AutoValidate)
	require.Equal(t, cfg.Paths.PathStrategy, decoded.Paths.PathStrategy)
	require.Equal(t, cfg.Globals.Env["AWS_REGION"], decoded.Globals.Env["AWS_REGION"])
}

func TestConfigPartialMergeWithDefaults(t *testing.T) {
	path := filepath.Join("testdata", "partial_config.toml")
	cfg, err := LoadFromPath(path)
	require.NoError(t, err)

	defaults := Default()
	require.Equal(t, ".env.custom", cfg.Defaults.EnvFile)
	require.Equal(t, defaults.General.WarnOnMissing, cfg.General.WarnOnMissing)
	require.Equal(t, defaults.Validation.SchemaFile, cfg.Validation.SchemaFile)
	require.Equal(t, defaults.Paths.PathStrategy, cfg.Paths.PathStrategy)
}

func TestConfigGlobalsEnvAccess(t *testing.T) {
	cfg := Default()
	cfg.Globals.Env = map[string]string{
		"EDITOR": "vim",
		"PAGER":  "less",
	}

	got := cfg.GetGlobalEnv()
	require.Equal(t, cfg.Globals.Env, got)

	got["EDITOR"] = "nano"
	require.Equal(t, "vim", cfg.Globals.Env["EDITOR"])
}

func TestConfigGlobalsEmpty(t *testing.T) {
	cfg := Default()
	cfg.Globals.Env = nil

	got := cfg.GetGlobalEnv()
	require.Empty(t, got)
}
