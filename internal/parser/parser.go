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

package parser

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Parser struct {
	Filename string
}

func NewParser(fn string) *Parser {
	return &Parser{Filename: fn}
}

func (p *Parser) Parse() (map[string]string, error) {
	if !p.fileExists(p.Filename) {
		return nil, fmt.Errorf("'%s': No such file or directory", p.Filename)
	}

	content := p.readFile(p.Filename)
	lines := p.readLines(content)
	variables := p.extractVariables(lines)

	return variables, nil
}

func (p *Parser) fileExists(fn string) bool {
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		return false
	}
	return true
}

func (p *Parser) readFile(fn string) string {
	dat, err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Printf("[Err] '%s': Permission denied\n", fn)
		os.Exit(-1)
	}
	return string(dat)
}

func (p *Parser) readLines(input string) []string {
	lines := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 { //ignore empty lines
			if !strings.HasPrefix(line, "#") { //ignore comments
				lines = append(lines, string(line))
			}
		}
	}
	return lines
}

func (p *Parser) extractVariables(lines []string) map[string]string {
	variables := make(map[string]string)

	for _, line := range lines {
		result := strings.Split(line, "=")
		if len(result) >= 2 {
			name := result[0]
			value := strings.Join(result[1:], "=")
			variables[name] = value
		}
	}

	return variables
}
