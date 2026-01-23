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
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/fatih/color"
	"github.com/jcouture/nv/internal/config"
)

func TestConfigureColors(t *testing.T) {
	orig := color.NoColor
	t.Cleanup(func() {
		color.NoColor = orig
	})

	t.Setenv("NO_COLOR", "1")
	configureColors(false)
	if !color.NoColor {
		t.Fatal("expected color.NoColor to be true when NO_COLOR is set")
	}
}

func TestNewRootCmdCommands(t *testing.T) {
	cmd := NewRootCmd("nv")
	names := map[string]bool{}
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, name := range []string{"run", "export", "print", "validate", "version"} {
		if !names[name] {
			t.Fatalf("expected command %q to be registered", name)
		}
	}
}

func TestNewRootCmdDefaultName(t *testing.T) {
	cmd := NewRootCmd("")
	if cmd.Use != "nv" {
		t.Fatalf("expected default name nv, got %q", cmd.Use)
	}
}

func TestExitError(t *testing.T) {
	err := exitError{code: 7}
	if !strings.Contains(err.Error(), "7") {
		t.Fatalf("unexpected error string: %s", err.Error())
	}
}

func TestExecuteInvalidCommand(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() {
		os.Args = origArgs
	})

	os.Args = []string{"nv", "does-not-exist"}
	exitCode := executeCommand("nv")

	if exitCode == 0 {
		t.Fatal("expected non-zero exit code for invalid command")
	}
}

func TestExecuteExitError(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() {
		os.Args = origArgs
	})

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cmd, args := exitCommand(7)
	os.Args = append([]string{"nv", "run", "--env-file", envFile, "--"}, append([]string{cmd}, args...)...)

	exitCode := executeCommand("nv")

	if exitCode != 7 {
		t.Fatalf("expected exit code 7, got %d", exitCode)
	}
}

func TestExecuteSuccess(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	t.Cleanup(func() {
		os.Args = origArgs
		exitFunc = origExit
	})

	os.Args = []string{"nv", "version", "--format", "text"}
	exitFunc = func(code int) {
		t.Fatalf("unexpected exit code %d", code)
	}

	Execute()
}

func TestRunVersion(t *testing.T) {
	tests := []struct {
		name       string
		opts       versionOptions
		wantErr    bool
		wantString string
		checkJSON  bool
	}{
		{
			name:       "short",
			opts:       versionOptions{short: true},
			wantErr:    false,
			wantString: Version,
		},
		{
			name:       "text",
			opts:       versionOptions{format: "text"},
			wantErr:    false,
			wantString: "nv version",
		},
		{
			name:      "json",
			opts:      versionOptions{format: "json"},
			wantErr:   false,
			checkJSON: true,
		},
		{
			name:    "invalid",
			opts:    versionOptions{format: "bad"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, _ := captureOutput(t, func() {
				err := runVersion(&tc.opts)
				if tc.wantErr && err == nil {
					t.Fatal("expected error")
				}
				if !tc.wantErr && err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})

			if tc.checkJSON {
				var payload map[string]string
				if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
					t.Fatalf("expected json output: %v", err)
				}
				if payload["version"] == "" {
					t.Fatalf("expected version in json output")
				}
				return
			}

			if tc.wantString != "" && !strings.Contains(stdout, tc.wantString) {
				t.Fatalf("stdout missing %q: %s", tc.wantString, stdout)
			}
		})
	}
}

func TestRunExport(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	opts := &exportOptions{
		envFiles: []string{envFile},
		format:   "shell",
	}

	stdout, _ := captureOutput(t, func() {
		if err := runExport(opts); err != nil {
			t.Fatalf("runExport error: %v", err)
		}
	})

	if !strings.Contains(stdout, "export FOO=\"bar\"") {
		t.Fatalf("unexpected output: %s", stdout)
	}
}

func TestRunExportMissingEnvFile(t *testing.T) {
	opts := &exportOptions{
		envFiles: []string{"missing.env"},
		format:   "shell",
	}

	if err := runExport(opts); err == nil {
		t.Fatal("expected error for missing env file")
	}
}

func TestRunValidateMissing(t *testing.T) {
	color.NoColor = true
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	schemaFile := filepath.Join(tmpDir, ".env.example")

	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if err := os.WriteFile(schemaFile, []byte("REQUIRED=present\n"), 0o644); err != nil {
		t.Fatalf("write schema file: %v", err)
	}

	opts := &validateOptions{
		envFiles:   []string{envFile},
		schemaFile: schemaFile,
	}

	_, stderr := captureOutput(t, func() {
		err := runValidate(opts)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	if !strings.Contains(stderr, "Missing required environment variables") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestRunValidateSuccess(t *testing.T) {
	color.NoColor = true
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	schemaFile := filepath.Join(tmpDir, ".env.example")

	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if err := os.WriteFile(schemaFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write schema file: %v", err)
	}

	opts := &validateOptions{
		envFiles:   []string{envFile},
		schemaFile: schemaFile,
		verbose:    true,
	}

	stdout, _ := captureOutput(t, func() {
		if err := runValidate(opts); err != nil {
			t.Fatalf("runValidate error: %v", err)
		}
	})

	if !strings.Contains(stdout, "Validation passed") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestRunValidateMissingEnvFile(t *testing.T) {
	opts := &validateOptions{
		envFiles:   []string{"missing.env"},
		schemaFile: defaultSchemaFile,
	}

	if err := runValidate(opts); err == nil {
		t.Fatal("expected error for missing env file")
	}
}

func TestValidateEnvironmentWarnings(t *testing.T) {
	color.NoColor = true
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	schemaFile := filepath.Join(tmpDir, ".env.example")

	if err := os.WriteFile(envFile, []byte("FOO=bar\nEXTRA=1\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if err := os.WriteFile(schemaFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write schema file: %v", err)
	}

	_, stderr := captureOutput(t, func() {
		err := validateEnvironment(map[string]string{
			"FOO":   "bar",
			"EXTRA": "1",
		}, validationOptions{
			schemaFile: schemaFile,
			strict:     true,
			envFiles:   []string{schemaFile},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(stderr, "schema file matches an env file") {
		t.Fatalf("expected circular schema warning, got: %s", stderr)
	}
	if !strings.Contains(stderr, "Environment variables not present in schema") {
		t.Fatalf("expected extra vars warning, got: %s", stderr)
	}
}

func TestValidateEnvironmentEmptySchema(t *testing.T) {
	color.NoColor = true
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, ".env.example")
	if err := os.WriteFile(schemaFile, []byte(""), 0o644); err != nil {
		t.Fatalf("write schema file: %v", err)
	}

	_, stderr := captureOutput(t, func() {
		err := validateEnvironment(map[string]string{}, validationOptions{
			schemaFile: schemaFile,
			envFiles:   []string{},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(stderr, "schema file is empty") {
		t.Fatalf("expected empty schema warning, got: %s", stderr)
	}
}

func TestValidateEnvironmentVerbose(t *testing.T) {
	color.NoColor = true
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, ".env.example")
	if err := os.WriteFile(schemaFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write schema file: %v", err)
	}

	stdout, _ := captureOutput(t, func() {
		err := validateEnvironment(map[string]string{
			"FOO": "bar",
		}, validationOptions{
			schemaFile: schemaFile,
			verbose:    true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(stdout, "Validation passed") {
		t.Fatalf("expected verbose success, got: %s", stdout)
	}
}

func TestRunRunExitCode(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cmd, args := exitCommand(7)
	opts := &runOptions{
		envFiles: []string{envFile},
	}

	cmdFlags := newRunCmdForTest(t, map[string]string{
		"env-file": envFile,
		"cascade":  "false",
	})
	err := runRun(cmdFlags, opts, append([]string{cmd}, args...))
	exitErr, ok := err.(exitError)
	if !ok {
		t.Fatalf("expected exitError, got %T", err)
	}
	if exitErr.code != 7 {
		t.Fatalf("expected exit code 7, got %d", exitErr.code)
	}
}

func TestRunRunValidateDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()
	envFile := filepath.Join(tmpDir, ".env")
	schemaFile := filepath.Join(tmpDir, ".env.example")

	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if err := os.WriteFile(schemaFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write schema file: %v", err)
	}

	opts := &runOptions{
		envFiles:   []string{envFile},
		dryRun:     true,
		validate:   true,
		schemaFile: schemaFile,
		format:     "shell",
	}

	stdout, _ := captureOutput(t, func() {
		cmdFlags := newRunCmdForTest(t, map[string]string{
			"env-file":    envFile,
			"cascade":     "false",
			"dry-run":     "true",
			"validate":    "true",
			"schema-file": schemaFile,
		})
		if err := runRun(cmdFlags, opts, []string{"echo", "ok"}); err != nil {
			t.Fatalf("runRun error: %v", err)
		}
	})

	if !strings.Contains(stdout, "export FOO=\"bar\"") {
		t.Fatalf("unexpected output: %s", stdout)
	}
}

func TestRunRunValidateMissingSchema(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	opts := &runOptions{
		envFiles:   []string{envFile},
		validate:   true,
		schemaFile: filepath.Join(tmpDir, "missing.env.example"),
	}

	cmdFlags := newRunCmdForTest(t, map[string]string{
		"env-file": envFile,
		"cascade":  "false",
		"validate": "true",
	})
	err := runRun(cmdFlags, opts, []string{"echo", "ok"})
	if err == nil {
		t.Fatal("expected error for missing schema file")
	}
}

func TestRunRunRespectsSchemaFlags(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()
	envFile := filepath.Join(tmpDir, ".env")
	schemaFile := filepath.Join(tmpDir, ".env.schema")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if err := os.WriteFile(schemaFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write schema file: %v", err)
	}

	opts := &runOptions{
		envFiles:   []string{envFile},
		dryRun:     true,
		validate:   true,
		schemaFile: schemaFile,
		format:     "shell",
	}

	stdout, _ := captureOutput(t, func() {
		cmdFlags := newRunCmdForTest(t, map[string]string{
			"env-file":      envFile,
			"dry-run":       "true",
			"validate":      "true",
			"schema-file":   schemaFile,
			"schema-strict": "true",
			"cascade":       "false",
		})
		if err := runRun(cmdFlags, opts, []string{"echo", "ok"}); err != nil {
			t.Fatalf("runRun error: %v", err)
		}
	})

	if !strings.Contains(stdout, "export FOO=\"bar\"") {
		t.Fatalf("unexpected output: %s", stdout)
	}
}

func TestRunRunInvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	opts := &runOptions{
		envFiles: []string{envFile},
		dryRun:   true,
		format:   "nope",
	}

	cmdFlags := newRunCmdForTest(t, map[string]string{
		"env-file": envFile,
		"cascade":  "false",
		"dry-run":  "true",
	})
	err := runRun(cmdFlags, opts, []string{"echo", "ok"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestRunRunMissingCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	opts := &runOptions{
		envFiles: []string{envFile},
	}

	cmdFlags := newRunCmdForTest(t, map[string]string{
		"env-file": envFile,
		"cascade":  "false",
	})
	err := runRun(cmdFlags, opts, []string{"does-not-exist"})
	if err == nil {
		t.Fatal("expected error for missing command")
	}
}

func TestLoadEnvironmentCascade(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("BASE=1\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.production"), []byte("ENV=prod\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	env, err := loadEnvironment(envOptions{
		cascade: true,
		env:     "production",
	})
	if err != nil {
		t.Fatalf("loadEnvironment error: %v", err)
	}
	if env["BASE"] != "1" || env["ENV"] != "prod" {
		t.Fatalf("unexpected env: %v", env)
	}
}

func TestLoadEnvironmentOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	opts := envOptions{
		envFiles:  []string{envFile},
		overrides: []string{"FOO=baz"},
	}

	env, err := loadEnvironment(opts)
	if err != nil {
		t.Fatalf("loadEnvironment error: %v", err)
	}
	if env["FOO"] != "baz" {
		t.Fatalf("FOO=%q, want baz", env["FOO"])
	}
}

func TestLoadEnvironmentGlobalsPriorityFirst(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=local\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	env, err := loadEnvironment(envOptions{
		envFiles: []string{envFile},
		globals:  map[string]string{"FOO": "global", "BAR": "global"},
		priority: config.GlobalsPriorityFirst,
	})
	if err != nil {
		t.Fatalf("loadEnvironment error: %v", err)
	}
	if env["FOO"] != "local" {
		t.Fatalf("FOO=%q, want local", env["FOO"])
	}
	if env["BAR"] != "global" {
		t.Fatalf("BAR=%q, want global", env["BAR"])
	}
}

func TestLoadEnvironmentGlobalsPriorityLast(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=local\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	env, err := loadEnvironment(envOptions{
		envFiles: []string{envFile},
		globals:  map[string]string{"FOO": "global"},
		priority: config.GlobalsPriorityLast,
	})
	if err != nil {
		t.Fatalf("loadEnvironment error: %v", err)
	}
	if env["FOO"] != "global" {
		t.Fatalf("FOO=%q, want global", env["FOO"])
	}
}

func TestLoadEnvironmentAutoLocal(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	localFile := filepath.Join(tmpDir, ".env.local")
	if err := os.WriteFile(envFile, []byte("FOO=base\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if err := os.WriteFile(localFile, []byte("FOO=local\n"), 0o644); err != nil {
		t.Fatalf("write env local file: %v", err)
	}

	env, err := loadEnvironment(envOptions{
		envFiles:  []string{envFile},
		autoLocal: true,
	})
	if err != nil {
		t.Fatalf("loadEnvironment error: %v", err)
	}
	if env["FOO"] != "local" {
		t.Fatalf("FOO=%q, want local", env["FOO"])
	}
}

func TestLoadEnvironmentInvalidOverride(t *testing.T) {
	opts := envOptions{
		envFiles:  []string{defaultEnvFile},
		overrides: []string{"NOVAL"},
	}

	_, err := loadEnvironment(opts)
	if err == nil {
		t.Fatal("expected error for invalid override")
	}
}

func TestParseOverrideMissingKey(t *testing.T) {
	if _, _, err := parseOverride("=value"); err == nil {
		t.Fatal("expected error for missing key")
	}
}

func exitCommand(code int) (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", "exit", strconv.Itoa(code)}
	}
	return "sh", []string{"-c", "exit " + strconv.Itoa(code)}
}

func TestHasCircularSchemaFalse(t *testing.T) {
	if hasCircularSchema(nil, "") {
		t.Fatal("expected false for empty inputs")
	}
}

func TestValidateEnvironmentMissingSchema(t *testing.T) {
	err := validateEnvironment(map[string]string{}, validationOptions{
		schemaFile: "missing.env.example",
	})
	if err == nil {
		t.Fatal("expected error for missing schema file")
	}
}
