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

package loader

import (
	"errors"
	"os"

	"github.com/jcouture/nv/internal/parser"
)

type Loader struct {
	preserve []string
	strict   bool
}

type Option func(*Loader)

func WithPreserve(vars []string) Option {
	return func(l *Loader) {
		l.preserve = vars
	}
}

func WithStrict(strict bool) Option {
	return func(l *Loader) {
		l.strict = strict
	}
}

func New(opts ...Option) *Loader {
	l := &Loader{
		preserve: []string{"PATH", "HOME", "USER"},
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *Loader) LoadFiles(files ...string) (map[string]string, error) {
	env := l.preservedEnv()
	for _, file := range files {
		if err := l.loadFile(file, env, false); err != nil {
			return nil, err
		}
	}
	return env, nil
}

func (l *Loader) preservedEnv() map[string]string {
	env := make(map[string]string)
	for _, key := range l.preserve {
		if val, ok := os.LookupEnv(key); ok {
			env[key] = val
		}
	}
	return env
}

func (l *Loader) loadFile(path string, env map[string]string, optional bool) error {
	opts := []parser.Option{
		parser.WithExistingEnv(env),
	}
	if l.strict {
		opts = append(opts, parser.WithStrictInterpolation())
	}

	parsed, err := parser.ParseFile(path, opts...)
	if err != nil {
		if optional && errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	for k, v := range parsed {
		env[k] = v
	}
	return nil
}
