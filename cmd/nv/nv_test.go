// Copyright 2015-2023 Jean-Philippe Couture
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
	"os"
	"testing"

	"github.com/jcouture/nv/internal/build"
	"github.com/stretchr/testify/assert"
)

func TestPrintVersion(t *testing.T) {
	originalVersion := build.Version
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	originalDeprecationNoticeEnabled := deprecationNoticeEnabled

	t.Cleanup(func() {
		build.Version = originalVersion
		os.Stdout = originalStdout
		os.Stderr = originalStderr
		deprecationNoticeEnabled = originalDeprecationNoticeEnabled
	})

	build.Version = "2.2.2"
	deprecationNoticeEnabled = func() bool { return true }

	stdoutReader, stdoutWriter, err := os.Pipe()
	assert.NoError(t, err)

	stderrReader, stderrWriter, err := os.Pipe()
	assert.NoError(t, err)

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	printVersion()

	assert.NoError(t, stdoutWriter.Close())
	assert.NoError(t, stderrWriter.Close())

	var stdout bytes.Buffer
	_, err = stdout.ReadFrom(stdoutReader)
	assert.NoError(t, err)

	var stderr bytes.Buffer
	_, err = stderr.ReadFrom(stderrReader)
	assert.NoError(t, err)

	assert.Equal(t, "nv version 2.2.2\n\n", stdout.String())
	assert.Equal(t, "*******************************************\nWARNING: nv v2.x is no longer maintained.\nUpgrade to v3:\n  brew uninstall nv\n  brew install --cask jcouture/tap/nv\nStay on v2: no action needed\n*******************************************\n", stderr.String())
}

func TestPrintDeprecationNoticeDisabled(t *testing.T) {
	originalStderr := os.Stderr
	originalDeprecationNoticeEnabled := deprecationNoticeEnabled

	t.Cleanup(func() {
		os.Stderr = originalStderr
		deprecationNoticeEnabled = originalDeprecationNoticeEnabled
	})

	deprecationNoticeEnabled = func() bool { return false }

	stderrReader, stderrWriter, err := os.Pipe()
	assert.NoError(t, err)

	os.Stderr = stderrWriter

	printDeprecationNotice()

	assert.NoError(t, stderrWriter.Close())

	var stderr bytes.Buffer
	_, err = stderr.ReadFrom(stderrReader)
	assert.NoError(t, err)

	assert.Empty(t, stderr.String())
}
