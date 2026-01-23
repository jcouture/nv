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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jcouture/nv/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage nvx configuration",
	}

	cmd.AddCommand(
		newConfigInitCmd(),
		newConfigShowCmd(),
		newConfigPathCmd(),
		newConfigValidateCmd(),
		newConfigEditCmd(),
		newConfigResetCmd(),
		newConfigMigrateCmd(),
		newConfigGetCmd(),
		newConfigSetCmd(),
		newConfigGlobalsCmd(),
	)

	return cmd
}

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create default config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			exists, err := config.ConfigExists()
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("config file already exists")
			}

			cfg := config.Default()
			if err := cfg.Save(); err != nil {
				return err
			}
			path, err := config.GetConfigPath()
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Created config file at: %s\n", path)
			return nil
		},
	}
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			fmt.Fprintln(os.Stdout, "Current configuration (merged with defaults):")
			if err := toml.NewEncoder(os.Stdout).Encode(cfg); err != nil {
				return err
			}
			path, err := config.GetConfigPath()
			if err == nil {
				fmt.Fprintf(os.Stdout, "Config file: %s\n", path)
			}
			return nil
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.GetConfigPath()
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, path)
			return nil
		},
	}
}

func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			exists, err := config.ConfigExists()
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("config file does not exist")
			}

			path, err := config.GetConfigPath()
			if err != nil {
				return err
			}
			cfg := config.Default()
			if _, err := toml.DecodeFile(path, cfg); err != nil {
				return err
			}

			errs := cfg.Validate()
			if len(errs) == 0 {
				fmt.Fprintln(os.Stdout, "Config is valid")
				return nil
			}
			for _, err := range errs {
				fmt.Fprintf(os.Stderr, "- %v\n", err)
			}
			return exitError{code: 1}
		},
	}
}

func newConfigEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit config file in $EDITOR",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.GetConfigPath()
			if err != nil {
				return err
			}
			exists, err := config.ConfigExists()
			if err != nil {
				return err
			}
			if !exists {
				cfg := config.Default()
				if err := cfg.SaveToPath(path); err != nil {
					return err
				}
			}

			editor := os.Getenv("EDITOR")
			if strings.TrimSpace(editor) == "" {
				return fmt.Errorf("$EDITOR is not set")
			}

			parts := strings.Fields(editor)
			if len(parts) == 0 {
				return fmt.Errorf("invalid $EDITOR value")
			}
			cmdEditor := exec.Command(parts[0], append(parts[1:], path)...) // #nosec G204 -- editor command is explicitly provided by the user via $EDITOR.
			cmdEditor.Stdin = os.Stdin
			cmdEditor.Stdout = os.Stdout
			cmdEditor.Stderr = os.Stderr
			return cmdEditor.Run()
		},
	}
}

func newConfigResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Delete config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.GetConfigPath()
			if err != nil {
				return err
			}
			if err := os.Remove(path); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("config file does not exist")
				}
				return err
			}
			fmt.Fprintln(os.Stdout, "Config file removed")
			return nil
		},
	}
}

func newConfigMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate legacy ~/.nv to config",
		RunE: func(cmd *cobra.Command, args []string) error {
			configExists, err := config.ConfigExists()
			if err != nil {
				return err
			}
			if configExists {
				return fmt.Errorf("config file already exists")
			}
			legacyExists, err := config.LegacyEnvExists()
			if err != nil {
				return err
			}
			if !legacyExists {
				return fmt.Errorf("legacy ~/.nv file not found")
			}

			_, err = config.MigrateLegacyEnv()
			return err
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			value, err := getConfigValue(cfg, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, value)
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfigForWrite()
			if err != nil {
				return err
			}
			if err := setConfigValue(cfg, args[0], args[1]); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Set %s\n", args[0])
			return nil
		},
	}
}

func newConfigGlobalsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "globals",
		Short: "Manage global environment variables",
	}

	cmd.AddCommand(
		newConfigGlobalsListCmd(),
		newConfigGlobalsSetCmd(),
		newConfigGlobalsUnsetCmd(),
		newConfigGlobalsClearCmd(),
	)

	return cmd
}

func newConfigGlobalsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List global variables",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			keys := make([]string, 0, len(cfg.Globals.Env))
			for key := range cfg.Globals.Env {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Fprintf(os.Stdout, "%s=%s\n", key, cfg.Globals.Env[key])
			}
			return nil
		},
	}
}

func newConfigGlobalsSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a global variable",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfigForWrite()
			if err != nil {
				return err
			}
			if cfg.Globals.Env == nil {
				cfg.Globals.Env = map[string]string{}
			}
			cfg.Globals.Env[args[0]] = args[1]
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Set %s=%s\n", args[0], args[1])
			return nil
		},
	}
}

func newConfigGlobalsUnsetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unset <key>",
		Short: "Remove a global variable",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfigForWrite()
			if err != nil {
				return err
			}
			delete(cfg.Globals.Env, args[0])
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Removed %s\n", args[0])
			return nil
		},
	}
}

func newConfigGlobalsClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Remove all global variables",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfigForWrite()
			if err != nil {
				return err
			}
			cfg.Globals.Env = map[string]string{}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "All global variables removed")
			return nil
		},
	}
}

func loadConfigForWrite() (*config.Config, error) {
	exists, err := config.ConfigExists()
	if err != nil {
		return nil, err
	}
	if !exists {
		return config.Default(), nil
	}
	path, err := config.GetConfigPath()
	if err != nil {
		return nil, err
	}
	return config.LoadFromPath(path)
}

func getConfigValue(cfg *config.Config, key string) (string, error) {
	switch key {
	case "general.verbosity":
		return strconv.Itoa(cfg.General.Verbosity), nil
	case "defaults.env_file":
		return cfg.Defaults.EnvFile, nil
	case "defaults.auto_local":
		return strconv.FormatBool(cfg.Defaults.AutoLocal), nil
	case "defaults.cascade":
		return strconv.FormatBool(cfg.Defaults.Cascade), nil
	case "defaults.dry_run":
		return strconv.FormatBool(cfg.Defaults.DryRun), nil
	case "validation.enabled":
		return strconv.FormatBool(cfg.Validation.Enabled), nil
	case "validation.schema_file":
		return cfg.Validation.SchemaFile, nil
	case "validation.strict":
		return strconv.FormatBool(cfg.Validation.Strict), nil
	case "globals.priority":
		return cfg.Globals.Priority, nil
	default:
		if strings.HasPrefix(key, "globals.env.") {
			name := strings.TrimPrefix(key, "globals.env.")
			return cfg.Globals.Env[name], nil
		}
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

func setConfigValue(cfg *config.Config, key string, value string) error {
	switch key {
	case "general.verbosity":
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		cfg.General.Verbosity = parsed
	case "defaults.env_file":
		cfg.Defaults.EnvFile = value
	case "defaults.auto_local":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.Defaults.AutoLocal = parsed
	case "defaults.cascade":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.Defaults.Cascade = parsed
	case "defaults.dry_run":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.Defaults.DryRun = parsed
	case "validation.enabled":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.Validation.Enabled = parsed
	case "validation.schema_file":
		cfg.Validation.SchemaFile = value
	case "validation.strict":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.Validation.Strict = parsed
	case "globals.priority":
		cfg.Globals.Priority = value
	default:
		if strings.HasPrefix(key, "globals.env.") {
			name := strings.TrimPrefix(key, "globals.env.")
			if cfg.Globals.Env == nil {
				cfg.Globals.Env = map[string]string{}
			}
			cfg.Globals.Env[name] = value
			return nil
		}
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}
