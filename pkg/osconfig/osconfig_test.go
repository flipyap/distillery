package osconfig_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ekristen/distillery/pkg/osconfig"
)

func TestOS_GetOS(t *testing.T) {
	tests := []struct {
		name     string
		os       *osconfig.OS
		expected []string
	}{
		{
			name:     "Windows",
			os:       osconfig.New(osconfig.Windows, osconfig.AMD64),
			expected: []string{"windows", "win"},
		},
		{
			name:     "Linux",
			os:       osconfig.New(osconfig.Linux, osconfig.ARM64),
			expected: []string{"linux"},
		},
		{
			name:     "Darwin",
			os:       osconfig.New(osconfig.Darwin, osconfig.AMD64),
			expected: []string{"darwin", "macos", "sonoma"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.os.GetOS())
			assert.ElementsMatch(t, tt.expected, tt.os.GetOS())
		})
	}
}

func TestOS_GetArchitectures(t *testing.T) {
	tests := []struct {
		name     string
		os       *osconfig.OS
		expected []string
	}{
		{
			name:     "Windows AMD64",
			os:       osconfig.New(osconfig.Windows, osconfig.AMD64),
			expected: []string{"amd64", "x86_64", "64bit", "64", "x86", "64-bit", "x86-64"},
		},
		{
			name:     "Linux ARM64",
			os:       osconfig.New(osconfig.Linux, osconfig.ARM64),
			expected: []string{"arm64", "aarch64", "armv8-a", "arm64-bit"},
		},
		{
			name:     "Darwin Universal",
			os:       osconfig.New(osconfig.Darwin, osconfig.AMD64),
			expected: []string{"amd64", "x86_64", "64bit", "64", "x86", "64-bit", "x86-64", "universal"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, tt.expected, tt.os.GetArchitectures())
		})
	}
}

func TestOS_GetExtensions(t *testing.T) {
	tests := []struct {
		name     string
		os       *osconfig.OS
		expected []string
	}{
		{
			name:     "Windows",
			os:       osconfig.New(osconfig.Windows, osconfig.AMD64),
			expected: []string{".exe"},
		},
		{
			name:     "Linux",
			os:       osconfig.New(osconfig.Linux, osconfig.ARM64),
			expected: []string{".AppImage"},
		},
		{
			name:     "Darwin",
			os:       osconfig.New(osconfig.Darwin, osconfig.AMD64),
			expected: []string{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, tt.expected, tt.os.GetExtensions())
		})
	}
}
