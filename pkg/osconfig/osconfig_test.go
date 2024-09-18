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

func TestOS_GetArchitecture(t *testing.T) {
	tests := []struct {
		name     string
		os       *osconfig.OS
		expected string
	}{
		{
			name:     "Windows AMD64",
			os:       osconfig.New(osconfig.Windows, osconfig.AMD64),
			expected: "amd64",
		},
		{
			name:     "Linux ARM64",
			os:       osconfig.New(osconfig.Linux, osconfig.ARM64),
			expected: "arm64",
		},
		{
			name:     "Darwin Universal",
			os:       osconfig.New(osconfig.Darwin, osconfig.AMD64),
			expected: "amd64",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.os.GetArchitecture())
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
			expected: []string{"amd64", "x86_64", "64bit", "x64", "x86", "64-bit", "x86-64"},
		},
		{
			name:     "Linux ARM64",
			os:       osconfig.New(osconfig.Linux, osconfig.ARM64),
			expected: []string{"arm64", "aarch64", "armv8-a", "arm64-bit"},
		},
		{
			name:     "Darwin Universal",
			os:       osconfig.New(osconfig.Darwin, osconfig.AMD64),
			expected: []string{"amd64", "x86_64", "64bit", "x64", "x86", "64-bit", "x86-64", "universal"},
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

func TestOS_InvalidOS(t *testing.T) {
	tests := []struct {
		name     string
		os       *osconfig.OS
		expected []string
	}{
		{
			name:     "Windows",
			os:       osconfig.New(osconfig.Windows, osconfig.AMD64),
			expected: []string{osconfig.Linux, osconfig.Darwin},
		},
		{
			name:     "Linux",
			os:       osconfig.New(osconfig.Linux, osconfig.ARM64),
			expected: []string{osconfig.Windows, osconfig.Darwin},
		},
		{
			name:     "Darwin",
			os:       osconfig.New(osconfig.Darwin, osconfig.AMD64),
			expected: []string{osconfig.Windows, osconfig.Linux},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.ElementsMatch(t, tc.expected, tc.os.InvalidOS())
		})
	}
}

func TestOS_InvalidArchitectures(t *testing.T) {
	tests := []struct {
		name     string
		os       *osconfig.OS
		expected []string
	}{
		{
			name:     "Windows AMD64",
			os:       osconfig.New(osconfig.Windows, osconfig.AMD64),
			expected: osconfig.ARM64Architectures,
		},
		{
			name:     "Linux ARM64",
			os:       osconfig.New(osconfig.Linux, osconfig.ARM64),
			expected: osconfig.AMD64Architectures,
		},
		{
			name:     "Darwin Universal",
			os:       osconfig.New(osconfig.Darwin, osconfig.AMD64),
			expected: osconfig.ARM64Architectures,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.os.InvalidArchitectures())
		})
	}
}

// test invalid os and architectures
func TestOS_InvalidOSArchitectures(t *testing.T) {
	os1 := osconfig.New("fake", "star")
	assert.Equal(t, []string{}, os1.InvalidOS())
	assert.Equal(t, []string{}, os1.InvalidArchitectures())
}
