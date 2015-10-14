// Copyright (c) 2015 Jean-Philippe Couture
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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("nv .env command [args]")
		return
	}
	fn := os.Args[1]
	cmd := os.Args[2]
	args := os.Args[2:]

	setEnvVars(loadEnvFile(fn))

	binary, lookErr := exec.LookPath(cmd)
	if lookErr != nil {
		binary = cmd
	}

	execErr := syscall.Exec(binary, args, os.Environ())
	if execErr != nil {
		fmt.Println("[Err] Cannot execute:", binary)
	}
}

func loadEnvFile(fn string) []string {
	dat, err := ioutil.ReadFile(fn)
	if err != nil {
		panic(err)
	}
	vars := make([]string, 0)
	lines := bytes.Split(dat, []byte("\n"))
	for _, line := range lines {
		if len(line) > 0 {
			vars = append(vars, string(line))
		}
	}
	return vars
}

func setEnvVars(envVars []string) {
	for _, envVar := range envVars {
		result := strings.Split(envVar, "=")
		if len(result) >= 2 {
			name := result[0]
			value := strings.Join(result[1:], "=")
			os.Setenv(name, value)
		}
	}
}
