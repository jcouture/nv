// Copyright 2015-2021 Jean-Philippe Couture
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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/jcouture/nv/internal/build"
	"github.com/jcouture/nv/internal/parser"
	"github.com/mitchellh/go-homedir"
)

func main() {
	if len(os.Args) == 2 {
		cmd := os.Args[1]
		switch cmd {
		case "-v", "version", "-version", "--version":
			printVersion()
		default:
			printHelp()
		}
		os.Exit(0)
	}

	if len(os.Args) < 3 {
		printHelp()
		os.Exit(0)
	}

	fn := os.Args[1]
	cmd := os.Args[2]
	args := os.Args[2:]

	filenames := strings.Split(fn, ",")

	vars := make(map[string]string)

	for _, filename := range filenames {
		// Parse file
		parser := parser.NewParser(filename)
		parsedVars, err := parser.Parse()
		if err != nil {
			fmt.Printf("[Err] %s\n", err)
			os.Exit(-1)
		}
		// Merge with possibly existing variables
		mergeVars(vars, parsedVars)
	}

	loadAndMergeGlobalVars(vars)

	clearEnv()
	setEnvVars(vars)

	binary, lookErr := exec.LookPath(cmd)
	if lookErr != nil {
		binary = cmd
	}

	execErr := syscall.Exec(binary, args, os.Environ())
	if execErr != nil {
		fmt.Println("[Err] Cannot execute:", binary)
		os.Exit(-1)
	}
}

func printHelp() {
	fmt.Printf("usage: nv <env files> <command> [arguments...]\n")
}

func printVersion() {
	fmt.Printf("nv version %s\n", build.Version)
}

func setEnvVars(vars map[string]string) {
	for k, v := range vars {
		os.Setenv(k, v)
	}
}

func mergeVars(vars1 map[string]string, vars2 map[string]string) {
	for k, v := range vars2 {
		vars1[k] = v
	}
}

func loadAndMergeGlobalVars(vars map[string]string) {
	dir, _ := homedir.Dir()
	fn := filepath.Join(dir, ".nv")
	parser := parser.NewParser(fn)
	parsedVars, err := parser.Parse()
	if err != nil {
		// Return without breaking a sweat
		return
	}
	mergeVars(vars, parsedVars)
}

func clearEnv() {
	// Clearing everything out the environment... but $PATH (we’re savages)!
	path := os.Getenv("PATH")
	os.Clearenv()
	os.Setenv("PATH", path)
}
