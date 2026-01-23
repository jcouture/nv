// Copyright (c) 2020-2025 Jonathan Couture
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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jcouture/nv/internal/parser"
)

var isInteractiveStdin = func() (bool, error) {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}
	return (stdinInfo.Mode() & os.ModeCharDevice) != 0, nil
}

func MigrateLegacyEnv() (bool, error) {
	configExists, err := ConfigExists()
	if err != nil {
		return false, err
	}
	if configExists {
		return false, nil
	}
	legacyExists, err := LegacyEnvExists()
	if err != nil {
		return false, err
	}
	if !legacyExists {
		return false, nil
	}

	legacyPath, err := GetLegacyEnvPath()
	if err != nil {
		return false, err
	}

	globals, err := ParseLegacyEnvFile(legacyPath)
	if err != nil {
		return false, err
	}

	cfg := CreateConfigWithGlobals(globals)
	if err := cfg.Save(); err != nil {
		return false, err
	}

	if err := BackupLegacyEnv(); err != nil {
		return false, err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return false, err
	}
	if _, err := LoadFromPath(configPath); err != nil {
		return false, err
	}

	if err := DeleteLegacyEnv(); err != nil {
		return false, err
	}

	fmt.Fprintln(os.Stdout, "migrated ~/.nv -> config (finally)")
	fmt.Fprintf(os.Stdout, "globals now live at: %s\n", configPath)
	fmt.Fprintf(os.Stdout, "backup dropped at: %s\n\n", filepath.Join(filepath.Dir(configPath), "nv.backup"))
	fmt.Fprintln(os.Stdout, "peek: nv config show")
	fmt.Fprintln(os.Stdout, "edit: nv config edit    # or just open the file, we won't stop you")

	return true, nil
}

func DetectLegacyEnv() (bool, error) {
	return LegacyEnvExists()
}

func PromptMigration() (bool, error) {
	interactive, err := isInteractiveStdin()
	if err != nil {
		return false, err
	}
	if !interactive {
		return false, nil
	}

	legacyPath, err := GetLegacyEnvPath()
	if err != nil {
		return false, err
	}
	configPath, err := GetConfigPath()
	if err != nil {
		return false, err
	}
	configDir, err := GetConfigDir()
	if err != nil {
		return false, err
	}

	fmt.Fprintf(os.Stdout, "Warning: legacy global env file detected: %s\n\n", legacyPath)
	fmt.Fprintln(os.Stdout, "nv moved to XDG paths (sigh).")
	fmt.Fprintln(os.Stdout, "We'll stash your globals here:")
	fmt.Fprintf(os.Stdout, "%s\n\n", configPath)
	fmt.Fprintln(os.Stdout, "Your ~/.nv file will be:")
	fmt.Fprintln(os.Stdout, "- Parsed and imported into [globals] section")
	fmt.Fprintf(os.Stdout, "- Backed up to %s\n", filepath.Join(configDir, "nv.backup"))
	fmt.Fprintln(os.Stdout, "- Removed from your home directory")
	fmt.Fprint(os.Stdout, "\nMigrate now? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err == io.EOF {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" || input == "y" || input == "yes" {
		return true, nil
	}
	return false, nil
}

func ParseLegacyEnvFile(path string) (map[string]string, error) {
	env := currentEnv()
	return parser.ParseFile(path, parser.WithExistingEnv(env))
}

func CreateConfigWithGlobals(globals map[string]string) *Config {
	cfg := Default()
	cfg.Globals.Env = make(map[string]string, len(globals))
	for k, v := range globals {
		cfg.Globals.Env[k] = v
	}
	return cfg
}

func BackupLegacyEnv() error {
	legacyPath, err := GetLegacyEnvPath()
	if err != nil {
		return err
	}
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return err
	}

	src, err := os.Open(legacyPath) // #nosec G304 -- legacy path is resolved from user home directory.
	if err != nil {
		return err
	}
	defer src.Close()

	backupPath := filepath.Join(configDir, "nv.backup")
	dst, err := os.OpenFile(backupPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304 -- backup path is under the XDG config directory.
	if err != nil {
		return err
	}
	defer dst.Close()
	if err := os.Chmod(backupPath, 0o600); err != nil {
		return err
	}

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

func DeleteLegacyEnv() error {
	legacyPath, err := GetLegacyEnvPath()
	if err != nil {
		return err
	}
	return os.Remove(legacyPath)
}

func MigrateNonInteractive() error {
	configExists, err := ConfigExists()
	if err != nil {
		return err
	}
	if configExists {
		return nil
	}
	legacyExists, err := LegacyEnvExists()
	if err != nil {
		return err
	}
	if !legacyExists {
		return nil
	}
	_, err = MigrateLegacyEnv()
	return err
}

func currentEnv() map[string]string {
	env := make(map[string]string)
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}
