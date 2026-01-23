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
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const unknown = "unknown"

type versionInfo struct {
	GitVersion string `json:"gitVersion"`
	GitCommit  string `json:"gitCommit"`
	GoVersion  string `json:"goVersion"`
	Compiler   string `json:"compiler"`
	Platform   string `json:"platform"`
}

var (
	gitVersion = Version
	gitCommit  = Commit

	versionInfoOnce sync.Once
	info            versionInfo
)

type versionOptions struct {
	short  bool
	format string
}

func getBuildInfo() *debug.BuildInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}

	return bi
}

func getGitVersion(bi *debug.BuildInfo) string {
	if bi == nil {
		return gitVersion
	}

	if bi.Main.Version == "(devel)" || bi.Main.Version == "" {
		return gitVersion
	}

	return bi.Main.Version
}

func getCommit(bi *debug.BuildInfo) string {
	return getKey(bi, "vcs.revision")
}

func getKey(bi *debug.BuildInfo, key string) string {
	if bi == nil {
		return unknown
	}

	for _, iter := range bi.Settings {
		if iter.Key == key {
			return iter.Value
		}
	}

	return unknown
}

func getVersionInfo() versionInfo {
	versionInfoOnce.Do(func() {
		bi := getBuildInfo()

		if gitVersion == "" || gitVersion == "dev" || gitVersion == unknown {
			gitVersion = getGitVersion(bi)
		}

		if gitCommit == "" || gitCommit == "none" || gitCommit == unknown {
			gitCommit = getCommit(bi)
		}

		info = versionInfo{
			GitVersion: gitVersion,
			GitCommit:  gitCommit,
			GoVersion:  runtime.Version(),
			Compiler:   runtime.Compiler,
			Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		}
	})

	return info
}

func (i versionInfo) String() string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintf(w, "GitVersion:\t%s\n", i.GitVersion)
	_, _ = fmt.Fprintf(w, "GitCommit:\t%s\n", i.GitCommit)
	_, _ = fmt.Fprintf(w, "GoVersion:\t%s\n", i.GoVersion)
	_, _ = fmt.Fprintf(w, "Compiler:\t%s\n", i.Compiler)
	_, _ = fmt.Fprintf(w, "Platform:\t%s\n", i.Platform)

	_ = w.Flush()

	return b.String()
}

func newVersionCmd() *cobra.Command {
	opts := &versionOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion(opts)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opts.short, "short", "s", false, "Print only version number")
	flags.StringVarP(&opts.format, "format", "f", "text", "Output format: text, json")

	return cmd
}

func runVersion(opts *versionOptions) error {
	info := getVersionInfo()

	if opts.short {
		fmt.Fprintln(os.Stdout, info.GitVersion)
		return nil
	}

	switch opts.format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(info)
	case "text":
		fmt.Fprintln(os.Stdout, banner())
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, info.String())
		return nil
	default:
		return fmt.Errorf("unknown format: %s", opts.format)
	}
}

func banner() string {
	lines := []string{
		"_ ____   __",
		"| '_ \\ \\ / /",
		"| | | \\ V / ",
		"|_| |_|\\_/",
	}

	palette := []color.Attribute{
		color.FgHiRed,
		color.FgHiMagenta,
		color.FgHiBlue,
		color.FgHiCyan,
		color.FgHiGreen,
		color.FgHiYellow,
	}

	colored := make([]string, 0, len(lines)+2)
	colored = append(colored, color.New(color.FgHiWhite).Sprint("❯ nv version"))
	for i, line := range lines {
		c := color.New(palette[i%len(palette)])
		colored = append(colored, c.Sprint(line))
	}
	colored = append(colored, color.New(color.FgHiWhite).Sprint("nv: Purpose-built env loader for your current task."))

	return strings.Join(colored, "\n")
}
