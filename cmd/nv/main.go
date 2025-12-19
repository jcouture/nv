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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/jcouture/nv/internal/build"
	"github.com/jcouture/nv/internal/sys"
	"github.com/jcouture/nv/pkg/env"
)

var (
	exitFunc     = os.Exit
	execFunc     = syscall.Exec
	lookPathFunc = exec.LookPath
	setvarsFunc  = env.Setvars
)

func main() {
	exitFunc(run(os.Args))
}

func run(args []string) int {
	if len(args) == 2 {
		cmd := args[1]
		switch cmd {
		case "-v", "version", "-version", "--version":
			printVersion()
		default:
			printHelp()
		}
		return 0
	}

	if len(args) < 3 {
		printHelp()
		return 0
	}

	filenames := strings.Split(args[1], ",")
	cmd := args[2]
	cmdArgs := args[2:]

	base := make(map[string]string)
	for _, filename := range filenames {
		override, err := sys.ReadVarsFromFile(filename)
		if err != nil {
			fmt.Printf("%s\n", err)
			return -1
		}
		base = env.Join(base, override)
	}

	base = env.Join(base, sys.ReadGlobalVars())

	env.Clear("PATH")
	if err := setvarsFunc(base); err != nil {
		fmt.Printf("%s\n", err)
		return -1
	}

	bin, err := lookPathFunc(cmd)
	if err != nil {
		bin = cmd
	}

	// #nosec G204 -- executing the user-specified command is the core behavior.
	if err := execFunc(bin, cmdArgs, os.Environ()); err != nil {
		fmt.Println("cannot execute:", bin)
		return -1
	}
	return 0
}

func printHelp() {
	fmt.Printf("usage: nv [--version] [--help]\n")
	fmt.Printf("          <env files> <command> [arguments...]\n")
}

func printVersion() {
	fmt.Printf("nv version %s\n", build.Version)
}
