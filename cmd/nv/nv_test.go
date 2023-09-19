package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadVars(t *testing.T) {
	cases := []struct {
		description string
		input       string
		expected    map[string]string
	}{
		{"Simple file", "testdata/.env", map[string]string{"PORT": "4200", "SECRET_KEY": "1234567890"}},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			result, err := loadVars(tt.input)
			if err != nil {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}
