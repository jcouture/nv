// Copyright 2015-2022 Jean-Philippe Couture
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

	"github.com/jcouture/env"
	"github.com/jcouture/nv/internal/build"
	"github.com/jcouture/nv/internal/sys"
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

	base := make(map[string]string)

	for _, filename := range filenames {
		override, err := sys.ReadVarsFromFile(filename)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(-1)
		}
		// Merge with possibly existing variables
		base = env.Join(base, override)
	}

	globals := sys.ReadGlobalVars()
	base = env.Join(base, globals)

	// Clearing everything out the environment... except $PATH (we’re savages)!
	env.Clear("PATH")
	env.Setvars(base)

	bin, err := exec.LookPath(cmd)
	if err != nil {
		bin = cmd
	}

	if err := syscall.Exec(bin, args, os.Environ()); err != nil {
		fmt.Println("cannot execute:", bin)
		os.Exit(-1)
	}
}

func printHelp() {
	fmt.Printf("usage: nv [--version] [--help]\n")
	fmt.Printf("          <env files> <command> [arguments...]\n")
}

func printVersion() {
	fmt.Printf("nv version %s\n", build.Version)
}
