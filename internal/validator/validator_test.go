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

package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateMissingRequired(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, ".env.example")
	schema := "DATABASE_URL=postgres://localhost\nOPTIONAL=\n# REQUIRED: API_KEY\n"
	if err := os.WriteFile(schemaPath, []byte(schema), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	env := map[string]string{
		"DATABASE_URL": "",
	}

	result, err := Validate(schemaPath, env, Options{})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}
	if len(result.Missing) != 2 {
		t.Fatalf("expected 2 missing keys, got %v", result.Missing)
	}
}

func TestValidateSuccessOptional(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, ".env.example")
	schema := "DATABASE_URL=postgres://localhost\nOPTIONAL=\n"
	if err := os.WriteFile(schemaPath, []byte(schema), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	env := map[string]string{
		"DATABASE_URL": "postgres://localhost",
	}

	result, err := Validate(schemaPath, env, Options{})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}
	if len(result.Missing) != 0 {
		t.Fatalf("expected no missing keys, got %v", result.Missing)
	}
}

func TestValidateStrictExtra(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, ".env.example")
	schema := "DATABASE_URL=postgres://localhost\n"
	if err := os.WriteFile(schemaPath, []byte(schema), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	env := map[string]string{
		"DATABASE_URL": "postgres://localhost",
		"EXTRA":        "1",
	}

	result, err := Validate(schemaPath, env, Options{Strict: true})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}
	if len(result.Extra) != 1 || result.Extra[0] != "EXTRA" {
		t.Fatalf("expected EXTRA to be reported, got %v", result.Extra)
	}
}

func TestValidateEmptySchema(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, ".env.example")
	if err := os.WriteFile(schemaPath, []byte(""), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	result, err := Validate(schemaPath, map[string]string{}, Options{})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}
	if !result.EmptySchema {
		t.Fatal("expected empty schema to be detected")
	}
}

func TestValidateMissingSchemaPath(t *testing.T) {
	_, err := Validate("", map[string]string{}, Options{})
	if err == nil {
		t.Fatal("expected error for empty schema path")
	}
}

func TestValidateMissingSchemaFile(t *testing.T) {
	_, err := Validate("does-not-exist.env", map[string]string{}, Options{})
	if err == nil {
		t.Fatal("expected error for missing schema file")
	}
}
