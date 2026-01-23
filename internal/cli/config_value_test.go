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

package cli

import (
	"testing"

	"github.com/jcouture/nv/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGetConfigValueGlobalsEnv(t *testing.T) {
	cfg := config.Default()
	cfg.Globals.Env["FOO"] = "bar"

	val, err := getConfigValue(cfg, "globals.env.FOO")
	require.NoError(t, err)
	require.Equal(t, "bar", val)
}

func TestSetConfigValueParseErrors(t *testing.T) {
	cfg := config.Default()

	err := setConfigValue(cfg, "general.verbosity", "nope")
	require.Error(t, err)

	err = setConfigValue(cfg, "general.auto_validate", "nope")
	require.Error(t, err)

	err = setConfigValue(cfg, "defaults.auto_local", "nope")
	require.Error(t, err)

	err = setConfigValue(cfg, "validation.strict", "nope")
	require.Error(t, err)

	err = setConfigValue(cfg, "validation.allow_extra", "nope")
	require.Error(t, err)

	err = setConfigValue(cfg, "defaults.dry_run", "nope")
	require.Error(t, err)
}

func TestSetConfigValueGlobalsEnv(t *testing.T) {
	cfg := config.Default()
	err := setConfigValue(cfg, "globals.env.NEW", "value")
	require.NoError(t, err)
	require.Equal(t, "value", cfg.Globals.Env["NEW"])
}

func TestSetConfigValueGlobalsEnvNilMap(t *testing.T) {
	cfg := config.Default()
	cfg.Globals.Env = nil
	err := setConfigValue(cfg, "globals.env.NEW", "value")
	require.NoError(t, err)
	require.Equal(t, "value", cfg.Globals.Env["NEW"])
}

func TestGetConfigValueGlobalsEnvMissing(t *testing.T) {
	cfg := config.Default()
	val, err := getConfigValue(cfg, "globals.env.MISSING")
	require.NoError(t, err)
	require.Equal(t, "", val)
}
func TestConfigValueRoundTrip(t *testing.T) {
	cfg := config.Default()

	updates := map[string]string{
		"general.auto_validate":   "true",
		"general.warn_on_missing": "false",
		"general.verbosity":       "2",
		"defaults.env_file":       ".env.custom",
		"defaults.auto_local":     "false",
		"defaults.cascade":        "false",
		"defaults.dry_run":        "true",
		"validation.enabled":      "true",
		"validation.schema_file":  ".env.schema",
		"validation.strict":       "false",
		"validation.allow_extra":  "false",
		"globals.priority":        "last",
	}

	for key, value := range updates {
		require.NoError(t, setConfigValue(cfg, key, value))
		got, err := getConfigValue(cfg, key)
		require.NoError(t, err)
		require.Equal(t, value, got)
	}
}
