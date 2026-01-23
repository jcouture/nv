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

type Config struct {
	General    GeneralConfig    `toml:"general"`
	Defaults   DefaultsConfig   `toml:"defaults"`
	Validation ValidationConfig `toml:"validation"`
	Globals    GlobalsConfig    `toml:"globals"`

	defined map[string]bool
}

var verbosityOverride int

type GeneralConfig struct {
	AutoValidate  bool `toml:"auto_validate"`
	WarnOnMissing bool `toml:"warn_on_missing"`
	Verbosity     int  `toml:"verbosity"`
}

type DefaultsConfig struct {
	EnvFile   string `toml:"env_file"`
	AutoLocal bool   `toml:"auto_local"`
	Cascade   bool   `toml:"cascade"`
	DryRun    bool   `toml:"dry_run"`
}

type ValidationConfig struct {
	Enabled    bool   `toml:"enabled"`
	SchemaFile string `toml:"schema_file"`
	Strict     bool   `toml:"strict"`
	AllowExtra bool   `toml:"allow_extra"`
}

type GlobalsConfig struct {
	Priority string            `toml:"priority"`
	Env      map[string]string `toml:"env,omitempty"`
}

const (
	GlobalsPriorityFirst = "first"
	GlobalsPriorityLast  = "last"
)

func (c *Config) GetGlobalEnv() map[string]string {
	if c == nil || len(c.Globals.Env) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(c.Globals.Env))
	for k, v := range c.Globals.Env {
		out[k] = v
	}
	return out
}

func SetVerbosityOverride(level int) {
	verbosityOverride = level
}

func ClearVerbosityOverride() {
	verbosityOverride = 0
}

func GetVerbosityOverride() int {
	return verbosityOverride
}
