// Copyright 2015-2018 Jean-Philippe Couture
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

	"github.com/jcouture/nv/parser"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(-1)
	}

	fn := os.Args[1]
	cmd := os.Args[2]
	args := os.Args[2:]

	filenames := strings.Split(fn, ",")

	vars := make(map[string]string)

	for _, filename := range filenames {
		// Parse file
		parser := parser.NewParser(filename)
		variables, err := parser.Parse()
		if err != nil {
			fmt.Printf("[Err] %s\n", err)
			os.Exit(-1)
		}
		// Merge with possibly existing variables
		for k, v := range variables {
			vars[k] = v
		}
	}

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

func printUsage() {
	usage := `nv - context specific environment variables
Usage: nv <env file(s)> <command> [arguments...]`
	fmt.Println(usage)
}

func setEnvVars(vars map[string]string) {
	for k, v := range vars {
		os.Setenv(k, v)
	}
}

func clearEnv() {
	// Clearing everything out the environment... but $PATH (we’re savages)!
	path := os.Getenv("PATH")
	os.Clearenv()
	os.Setenv("PATH", path)
}
