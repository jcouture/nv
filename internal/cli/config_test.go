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
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/jcouture/nv/internal/config"
	"github.com/stretchr/testify/require"
)

func TestConfigInitShowPathReset(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	initCmd := newConfigInitCmd()
	require.NoError(t, initCmd.RunE(initCmd, nil))

	pathCmd := newConfigPathCmd()
	stdout := captureStdout(t, func() {
		require.NoError(t, pathCmd.RunE(pathCmd, nil))
	})
	require.Contains(t, stdout, filepath.Join(tmpDir, "xdg", "nv", "config.toml"))

	showCmd := newConfigShowCmd()
	require.NoError(t, showCmd.RunE(showCmd, nil))

	resetCmd := newConfigResetCmd()
	require.NoError(t, resetCmd.RunE(resetCmd, nil))

	err := resetCmd.RunE(resetCmd, nil)
	require.Error(t, err)
}

func TestConfigPathError(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	pathCmd := newConfigPathCmd()
	err := pathCmd.RunE(pathCmd, nil)
	require.Error(t, err)
}

func TestNewConfigCmdHasSubcommands(t *testing.T) {
	cmd := newConfigCmd()
	if len(cmd.Commands()) == 0 {
		t.Fatalf("expected subcommands")
	}
}

func TestConfigInitWhenExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	cfg := config.Default()
	require.NoError(t, cfg.Save())

	initCmd := newConfigInitCmd()
	err := initCmd.RunE(initCmd, nil)
	require.Error(t, err)
}

func TestConfigInitWithBadConfigHome(t *testing.T) {
	tmpDir := t.TempDir()
	badHome := filepath.Join(tmpDir, "bad")
	require.NoError(t, os.WriteFile(badHome, []byte("nope"), 0o600))
	t.Setenv("XDG_CONFIG_HOME", badHome)
	xdg.Reload()

	initCmd := newConfigInitCmd()
	err := initCmd.RunE(initCmd, nil)
	require.Error(t, err)
}

func TestConfigInitSaveError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits behave differently on windows")
	}

	tmpDir := t.TempDir()
	readOnly := filepath.Join(tmpDir, "ro")
	require.NoError(t, os.MkdirAll(readOnly, 0o500))
	t.Setenv("XDG_CONFIG_HOME", readOnly)
	xdg.Reload()

	initCmd := newConfigInitCmd()
	err := initCmd.RunE(initCmd, nil)
	require.Error(t, err)
}

func TestConfigInitPathError(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	initCmd := newConfigInitCmd()
	err := initCmd.RunE(initCmd, nil)
	require.Error(t, err)
}

func TestConfigValidate(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	path, err := config.GetConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte("[globals]\npriority=\"bad\"\n"), 0o600))

	validateCmd := newConfigValidateCmd()
	err = validateCmd.RunE(validateCmd, nil)
	require.Error(t, err)
}

func TestConfigValidateValid(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	cfg := config.Default()
	require.NoError(t, cfg.Save())

	validateCmd := newConfigValidateCmd()
	err := validateCmd.RunE(validateCmd, nil)
	require.NoError(t, err)
}

func TestConfigValidateDecodeError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	path, err := config.GetConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte("[general]\nverbosity=\n"), 0o600))

	validateCmd := newConfigValidateCmd()
	err = validateCmd.RunE(validateCmd, nil)
	require.Error(t, err)
}

func TestConfigValidateMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	validateCmd := newConfigValidateCmd()
	err := validateCmd.RunE(validateCmd, nil)
	require.Error(t, err)
}

func TestConfigValidatePathError(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	validateCmd := newConfigValidateCmd()
	err := validateCmd.RunE(validateCmd, nil)
	require.Error(t, err)
}

func TestConfigEdit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("EDITOR handling is shell-specific on windows")
	}

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	os.Setenv("EDITOR", "true")
	xdg.Reload()

	editCmd := newConfigEditCmd()
	require.NoError(t, editCmd.RunE(editCmd, nil))
}

func TestConfigEditWhenExists(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("EDITOR handling is shell-specific on windows")
	}

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	os.Setenv("EDITOR", "true")
	xdg.Reload()

	cfg := config.Default()
	require.NoError(t, cfg.Save())

	editCmd := newConfigEditCmd()
	require.NoError(t, editCmd.RunE(editCmd, nil))
}

func TestConfigEditCommandFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("EDITOR handling is shell-specific on windows")
	}

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	os.Setenv("EDITOR", "false")
	xdg.Reload()

	editCmd := newConfigEditCmd()
	err := editCmd.RunE(editCmd, nil)
	require.Error(t, err)
}

func TestConfigEditSaveError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits behave differently on windows")
	}

	tmpDir := t.TempDir()
	readOnly := filepath.Join(tmpDir, "ro")
	require.NoError(t, os.MkdirAll(readOnly, 0o500))
	t.Setenv("XDG_CONFIG_HOME", readOnly)
	os.Setenv("EDITOR", "true")
	xdg.Reload()

	editCmd := newConfigEditCmd()
	err := editCmd.RunE(editCmd, nil)
	require.Error(t, err)
}

func TestConfigEditPathError(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	editCmd := newConfigEditCmd()
	err := editCmd.RunE(editCmd, nil)
	require.Error(t, err)
}

func TestConfigEditMissingEditor(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	os.Setenv("EDITOR", "")
	xdg.Reload()

	editCmd := newConfigEditCmd()
	err := editCmd.RunE(editCmd, nil)
	require.Error(t, err)
}

func TestConfigEditInvalidEditor(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	os.Setenv("EDITOR", "   ")
	xdg.Reload()

	editCmd := newConfigEditCmd()
	err := editCmd.RunE(editCmd, nil)
	require.Error(t, err)
}

func TestConfigGetSet(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	setCmd := newConfigSetCmd()
	setCmd.SetArgs([]string{"general.verbosity", "2"})
	require.NoError(t, setCmd.RunE(setCmd, []string{"general.verbosity", "2"}))

	getCmd := newConfigGetCmd()
	stdout := captureStdout(t, func() {
		require.NoError(t, getCmd.RunE(getCmd, []string{"general.verbosity"}))
	})
	require.Equal(t, "2\n", stdout)
}

func TestConfigGetCommandUnknownKey(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	getCmd := newConfigGetCmd()
	err := getCmd.RunE(getCmd, []string{"unknown.key"})
	require.Error(t, err)
}

func TestConfigSetCommandInvalidValue(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	setCmd := newConfigSetCmd()
	err := setCmd.RunE(setCmd, []string{"validation.enabled", "nope"})
	require.Error(t, err)
}

func TestConfigSetCommandSaveError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits behave differently on windows")
	}

	tmpDir := t.TempDir()
	readOnly := filepath.Join(tmpDir, "ro")
	require.NoError(t, os.MkdirAll(readOnly, 0o500))
	t.Setenv("XDG_CONFIG_HOME", readOnly)
	xdg.Reload()

	setCmd := newConfigSetCmd()
	err := setCmd.RunE(setCmd, []string{"validation.enabled", "true"})
	require.Error(t, err)
}

func TestConfigGlobalsCommands(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	setCmd := newConfigGlobalsSetCmd()
	require.NoError(t, setCmd.RunE(setCmd, []string{"AWS_REGION", "us-east-1"}))

	listCmd := newConfigGlobalsListCmd()
	stdout := captureStdout(t, func() {
		require.NoError(t, listCmd.RunE(listCmd, nil))
	})
	require.Contains(t, stdout, "AWS_REGION=us-east-1")

	unsetCmd := newConfigGlobalsUnsetCmd()
	require.NoError(t, unsetCmd.RunE(unsetCmd, []string{"AWS_REGION"}))

	clearCmd := newConfigGlobalsClearCmd()
	require.NoError(t, clearCmd.RunE(clearCmd, nil))
}

func TestConfigGlobalsUnsetMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	unsetCmd := newConfigGlobalsUnsetCmd()
	require.NoError(t, unsetCmd.RunE(unsetCmd, []string{"MISSING"}))
}

func TestConfigGetGlobalsEnv(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	cfg := config.Default()
	cfg.Globals.Env["FOO"] = "bar"
	require.NoError(t, cfg.Save())

	getCmd := newConfigGetCmd()
	stdout := captureStdout(t, func() {
		require.NoError(t, getCmd.RunE(getCmd, []string{"globals.env.FOO"}))
	})
	require.Equal(t, "bar\n", stdout)
}

func TestConfigGlobalsSetSaveError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits behave differently on windows")
	}

	tmpDir := t.TempDir()
	readOnly := filepath.Join(tmpDir, "ro")
	require.NoError(t, os.MkdirAll(readOnly, 0o500))
	t.Setenv("XDG_CONFIG_HOME", readOnly)
	xdg.Reload()

	setCmd := newConfigGlobalsSetCmd()
	err := setCmd.RunE(setCmd, []string{"FOO", "bar"})
	require.Error(t, err)
}

func TestConfigGlobalsSetPathError(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	setCmd := newConfigGlobalsSetCmd()
	err := setCmd.RunE(setCmd, []string{"FOO", "bar"})
	require.Error(t, err)
}

func TestConfigGlobalsUnsetSaveError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits behave differently on windows")
	}

	tmpDir := t.TempDir()
	readOnly := filepath.Join(tmpDir, "ro")
	require.NoError(t, os.MkdirAll(readOnly, 0o500))
	t.Setenv("XDG_CONFIG_HOME", readOnly)
	xdg.Reload()

	unsetCmd := newConfigGlobalsUnsetCmd()
	err := unsetCmd.RunE(unsetCmd, []string{"FOO"})
	require.Error(t, err)
}

func TestConfigGlobalsClearSaveError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits behave differently on windows")
	}

	tmpDir := t.TempDir()
	readOnly := filepath.Join(tmpDir, "ro")
	require.NoError(t, os.MkdirAll(readOnly, 0o500))
	t.Setenv("XDG_CONFIG_HOME", readOnly)
	xdg.Reload()

	clearCmd := newConfigGlobalsClearCmd()
	err := clearCmd.RunE(clearCmd, nil)
	require.Error(t, err)
}

func TestConfigGlobalsListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	listCmd := newConfigGlobalsListCmd()
	stdout := captureStdout(t, func() {
		require.NoError(t, listCmd.RunE(listCmd, nil))
	})
	require.Equal(t, "", stdout)
}

func TestConfigMigrateCmd(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	legacyPath := filepath.Join(tmpDir, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	migrateCmd := newConfigMigrateCmd()
	require.NoError(t, migrateCmd.RunE(migrateCmd, nil))
}

func TestConfigMigrateCmdMissingLegacy(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	migrateCmd := newConfigMigrateCmd()
	err := migrateCmd.RunE(migrateCmd, nil)
	require.Error(t, err)
}

func TestConfigMigrateCmdConfigExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	cfg := config.Default()
	require.NoError(t, cfg.Save())

	legacyPath := filepath.Join(tmpDir, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	migrateCmd := newConfigMigrateCmd()
	err := migrateCmd.RunE(migrateCmd, nil)
	require.Error(t, err)
}

func TestConfigMigrateCmdConfigExistsError(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	migrateCmd := newConfigMigrateCmd()
	err := migrateCmd.RunE(migrateCmd, nil)
	require.Error(t, err)
}

func TestConfigResetPathError(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	resetCmd := newConfigResetCmd()
	err := resetCmd.RunE(resetCmd, nil)
	require.Error(t, err)
}

func TestConfigResetPermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits behave differently on windows")
	}

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "xdg")
	t.Setenv("XDG_CONFIG_HOME", configDir)
	xdg.Reload()

	cfg := config.Default()
	require.NoError(t, cfg.Save())

	nvDir := filepath.Join(configDir, "nv")
	require.NoError(t, os.Chmod(nvDir, 0o500))
	t.Cleanup(func() {
		_ = os.Chmod(nvDir, 0o700)
	})

	resetCmd := newConfigResetCmd()
	err := resetCmd.RunE(resetCmd, nil)
	require.Error(t, err)
}

func TestConfigGetUnknownKey(t *testing.T) {
	cfg := config.Default()
	_, err := getConfigValue(cfg, "unknown.key")
	require.Error(t, err)
}

func TestConfigSetUnknownKey(t *testing.T) {
	cfg := config.Default()
	err := setConfigValue(cfg, "unknown.key", "value")
	require.Error(t, err)
}

func TestLoadConfigForWrite(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	cfg, err := loadConfigForWrite()
	require.NoError(t, err)
	require.Equal(t, config.Default().Defaults.EnvFile, cfg.Defaults.EnvFile)

	cfg.Defaults.EnvFile = ".env.custom"
	require.NoError(t, cfg.Save())

	cfg2, err := loadConfigForWrite()
	require.NoError(t, err)
	require.Equal(t, ".env.custom", cfg2.Defaults.EnvFile)
}

func TestLoadConfigForWriteError(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	_, err := loadConfigForWrite()
	require.Error(t, err)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = buf.ReadFrom(reader)
		close(done)
	}()

	fn()
	_ = writer.Close()
	<-done
	os.Stdout = origStdout

	return buf.String()
}

func TestConfigShowIncludesHeader(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	showCmd := newConfigShowCmd()
	stdout := captureStdout(t, func() {
		require.NoError(t, showCmd.RunE(showCmd, nil))
	})
	require.True(t, strings.Contains(stdout, "Current configuration"))
}

func TestConfigShowEncoderError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	showCmd := newConfigShowCmd()
	origStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	_ = writer.Close()
	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = origStdout
		_ = reader.Close()
	})

	err = showCmd.RunE(showCmd, nil)
	require.Error(t, err)
}
