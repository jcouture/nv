package env

import (
	"os"
	"sort"
	"testing"
)

func restoreEnv(t *testing.T, snapshot map[string]string) {
	t.Helper()
	os.Clearenv()
	for k, v := range snapshot {
		os.Setenv(k, v)
	}
}

func TestClear(t *testing.T) {
	snapshot := Getvars()
	defer restoreEnv(t, snapshot)

	const keepKey = "NV_TEST_KEEP"
	const dropKey = "NV_TEST_DROP"

	os.Setenv(keepKey, "keep")
	os.Setenv(dropKey, "drop")

	Clear(keepKey)

	if val, ok := os.LookupEnv(keepKey); !ok || val != "keep" {
		t.Fatalf("expected %s to be kept", keepKey)
	}
	if _, ok := os.LookupEnv(dropKey); ok {
		t.Fatalf("expected %s to be cleared", dropKey)
	}
}

func TestExists(t *testing.T) {
	const key = "NV_TEST_EXISTS"

	if Exists(key) {
		t.Fatalf("expected %s to not exist before setting", key)
	}

	os.Setenv(key, "1")
	defer os.Unsetenv(key)

	if !Exists(key) {
		t.Fatalf("expected %s to exist after setting", key)
	}
}

func TestGetvars(t *testing.T) {
	snapshot := Getvars()
	defer restoreEnv(t, snapshot)

	const key = "NV_TEST_GETVARS"
	os.Setenv(key, "value")

	vars := Getvars()
	if got := vars[key]; got != "value" {
		t.Fatalf("expected %s=value in vars, got %q", key, got)
	}
}

func TestGetnames(t *testing.T) {
	in := map[string]string{
		"A":    "1",
		"":     "ignored",
		"ZED":  "3",
		"NV_K": "2",
	}

	names := Getnames(in)
	sort.Strings(names)

	expected := []string{"A", "NV_K", "ZED"}
	if len(names) != len(expected) {
		t.Fatalf("expected names length %d, got %d", len(expected), len(names))
	}
	for i, n := range expected {
		if names[i] != n {
			t.Fatalf("expected names[%d]=%s, got %s", i, n, names[i])
		}
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]string
		override map[string]string
		want     map[string]string
	}{
		{
			name: "base empty",
			base: map[string]string{},
			override: map[string]string{
				"A": "1",
			},
			want: map[string]string{
				"A": "1",
			},
		},
		{
			name: "override wins",
			base: map[string]string{
				"A": "old",
				"B": "keep",
			},
			override: map[string]string{
				"A": "new",
				"C": "add",
			},
			want: map[string]string{
				"A": "new",
				"B": "keep",
				"C": "add",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := Join(tt.base, tt.override)
			if len(got) != len(tt.want) {
				t.Fatalf("expected len %d, got %d", len(tt.want), len(got))
			}
			for k, want := range tt.want {
				if got[k] != want {
					t.Fatalf("expected %s=%s, got %s", k, want, got[k])
				}
			}
		})
	}
}

func TestSetvars(t *testing.T) {
	snapshot := Getvars()
	defer restoreEnv(t, snapshot)

	input := map[string]string{
		"NV_TEST_ONE": "1",
		"NV_TEST_TWO": "2",
	}

	if err := Setvars(input); err != nil {
		t.Fatalf("Setvars returned error: %v", err)
	}

	for k, v := range input {
		if got := os.Getenv(k); got != v {
			t.Fatalf("expected %s=%s, got %s", k, v, got)
		}
	}
}
