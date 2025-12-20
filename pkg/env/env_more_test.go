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

package env

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExistsGetvarsGetnames(t *testing.T) {
	t.Setenv("NVX_ENV_TEST", "value")

	require.True(t, Exists("NVX_ENV_TEST"))
	require.False(t, Exists("NVX_ENV_MISSING"))

	vars := Getvars()
	require.Equal(t, "value", vars["NVX_ENV_TEST"])

	names := Getnames(map[string]string{"A": "1", "": "skip"})
	require.Contains(t, names, "A")
}

func TestJoinOverrides(t *testing.T) {
	base := map[string]string{"A": "1"}
	override := map[string]string{"A": "2", "B": "3"}

	merged := Join(base, override)
	require.Equal(t, "2", merged["A"])
	require.Equal(t, "3", merged["B"])
}

func TestSetvarsAndClear(t *testing.T) {
	vars := map[string]string{"NVX_SET_TEST": "1", "NVX_CLEAR_TEST": "2"}
	require.NoError(t, Setvars(vars))

	cleared := Clear("NVX_SET_TEST")
	require.NoError(t, cleared)
	require.True(t, Exists("NVX_SET_TEST"))
	require.False(t, Exists("NVX_CLEAR_TEST"))
}
