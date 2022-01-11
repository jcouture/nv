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
	"fmt"
	"testing"
)

func TestReadLines(t *testing.T) {
	input := `PORT=4200
	SECRET_KEY=1234567

	# This is comment
	DATABASE_URL=postgres://simonprev:@localhost:5432/accent_playground_dev?pool=10
	`

	p := new(Parser)

	lines := p.readLines(input)
	fmt.Printf("lines: %v\n", lines)
	actualSize := len(lines)
	expectedSize := 3
	if actualSize != expectedSize {
		t.Error("Expected: ", expectedSize, ", got: ", actualSize)
	}
}

func TestExtractVariables(t *testing.T) {
	input := make([]string, 3)
	input[0] = "PORT=4200"
	input[1] = "SECRET_KEY=1234567"
	input[2] = "DATABASE_URL=postgres://simonprev:@localhost:5432/accent_playground_dev?pool=10"

	p := new(Parser)

	variables := p.extractVariables(input)
	fmt.Printf("variables: %v\n", variables)
	actualSize := len(variables)
	expectedSize := 3
	if actualSize != expectedSize {
		t.Error("Expected: ", expectedSize, ", got: ", actualSize)
	}
}
