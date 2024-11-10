package inventory_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ekristen/distillery/pkg/config"
	"github.com/ekristen/distillery/pkg/inventory"
)

func TestInventory_New(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "inventory_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir)
	cfg, err := config.New("")
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	symPath := filepath.Join(tempDir, ".distillery", "bin")
	binPath := filepath.Join(tempDir, ".distillery", "opt")
	_ = os.MkdirAll(symPath, 0755)
	_ = os.MkdirAll(binPath, 0755)

	binaries := map[string]string{
		"test":               "github/ekristen/test/2.0.0/test",
		"test@1.0.0":         "github/ekristen/test/1.0.0/test",
		"test@2.0.0":         "github/ekristen/test/2.0.0/test",
		"another-test@1.0.0": "github/ekristen/another-test/1.0.0/another-test",
		"another-test@1.0.1": "github/ekristen/another-test/1.0.1/another-test",
	}

	for bin, target := range binaries {
		targetBase := filepath.Dir(target)
		targetName := filepath.Base(target)
		realBin := filepath.Join(binPath, targetBase, targetName)
		_ = os.MkdirAll(filepath.Join(realBin, targetBase), 0755)
		_ = os.WriteFile(realBin, []byte("test"), 0600)

		symlinkPath := filepath.Join(symPath, bin)
		if err := os.Symlink(realBin, symlinkPath); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}
	}

	dirFS := os.DirFS(tempDir)
	inv := inventory.New(dirFS, tempDir, ".distillery/bin", cfg)

	assert.NotNil(t, inv)
	assert.Equal(t, 2, inv.Count())
	assert.Equal(t, len(binaries), inv.FullCount())
}

func TestInventory_AddVersion(t *testing.T) {
	cases := []struct {
		name     string
		bins     map[string]string
		expected map[string]*inventory.Bin
	}{
		{
			name: "simple",
			bins: map[string]string{
				"/home/test/.distillery/bin/test@1.0.0": "/home/test/.distillery/opt/github/ekristen/test/1.0.0/test",
			},
			expected: map[string]*inventory.Bin{
				"test": {
					Name: "test",
					Versions: []*inventory.Version{
						{
							Version: "1.0.0",
							Path:    "/home/test/.distillery/bin/test@1.0.0",
							Target:  "/home/test/.distillery/opt/github/ekristen/test/1.0.0/test",
						},
					},
				},
			},
		},
		{
			name: "multiple",
			bins: map[string]string{
				"/home/test/.distillery/bin/test@1.0.0": "/home/test/.distillery/opt/github/ekristen/test/1.0.0/test",
				"/home/test/.distillery/bin/test@2.0.0": "/home/test/.distillery/opt/github/ekristen/test/2.0.0/test",
			},
			expected: map[string]*inventory.Bin{
				"test": {
					Name: "test",
					Versions: []*inventory.Version{
						{
							Version: "1.0.0",
							Path:    "/home/test/.distillery/bin/test@1.0.0",
							Target:  "/home/test/.distillery/opt/github/ekristen/test/1.0.0/test",
						},
						{
							Version: "2.0.0",
							Path:    "/home/test/.distillery/bin/test@2.0.0",
							Target:  "/home/test/.distillery/opt/github/ekristen/test/2.0.0/test",
						},
					},
				},
			},
		},
		{
			name: "complex",
			bins: map[string]string{
				"/home/test/.distillery/bin/test@1.0.0":         "/home/test/.distillery/opt/github/ekristen/test/1.0.0/test",
				"/home/test/.distillery/bin/test@1.0.1":         "/home/test/.distillery/opt/github/ekristen/test/1.0.1/test",
				"/home/test/.distillery/bin/another-test@1.0.0": "/home/test/.distillery/opt/github/ekristen/another-test/1.0.0/another-test",
				"/home/test/.distillery/bin/another-test@1.0.1": "/home/test/.distillery/opt/github/ekristen/another-test/1.0.1/another-test",
			},
			expected: map[string]*inventory.Bin{
				"test": {
					Name: "test",
					Versions: []*inventory.Version{
						{
							Version: "1.0.0",
							Path:    "/home/test/.distillery/bin/test@1.0.0",
							Target:  "/home/test/.distillery/opt/github/ekristen/test/1.0.0/test",
						},
						{
							Version: "1.0.1",
							Path:    "/home/test/.distillery/bin/test@1.0.1",
							Target:  "/home/test/.distillery/opt/github/ekristen/test/1.0.1/test",
						},
					},
				},
				"another-test": {
					Name: "another-test",
					Versions: []*inventory.Version{
						{
							Version: "1.0.0",
							Path:    "/home/test/.distillery/bin/another-test@1.0.0",
							Target:  "/home/test/.distillery/opt/github/ekristen/another-test/1.0.0/another-test",
						},
						{
							Version: "1.0.1",
							Path:    "/home/test/.distillery/bin/another-test@1.0.1",
							Target:  "/home/test/.distillery/opt/github/ekristen/another-test/1.0.1/another-test",
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inv := inventory.Inventory{}
			for bin, target := range tc.bins {
				_ = inv.AddVersion(bin, target)
			}

			assert.EqualValues(t, tc.expected, inv.Bins)
		})
	}
}

func BenchmarkInventoryNew(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "inventory_test")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir) // Clean up the temp directory after the test
	cfg, err := config.New("")
	if err != nil {
		b.Fatalf("Failed to create config: %v", err)
	}

	symPath := filepath.Join(tempDir, ".distillery", "bin")
	binPath := filepath.Join(tempDir, ".distillery", "opt")
	_ = os.MkdirAll(symPath, 0755)
	_ = os.MkdirAll(binPath, 0755)

	// Generate fake binary files to simulate version binaries
	binaries := map[string]string{
		"test":               "github/ekristen/test/2.0.0/test",
		"test@1.0.0":         "github/ekristen/test/1.0.0/test",
		"test@2.0.0":         "github/ekristen/test/2.0.0/test",
		"another-test@1.0.0": "github/ekristen/another-test/1.0.0/another-test",
		"another-test@1.0.1": "github/ekristen/another-test/1.0.1/another-test",
	}

	for bin, target := range binaries {
		targetBase := filepath.Dir(target)
		targetName := filepath.Base(target)
		realBin := filepath.Join(binPath, targetBase, targetName)
		_ = os.MkdirAll(filepath.Join(realBin, targetBase), 0755)
		_ = os.WriteFile(realBin, []byte("test"), 0600)

		symlinkPath := filepath.Join(symPath, bin)
		if err := os.Symlink(realBin, symlinkPath); err != nil {
			b.Fatalf("Failed to create symlink: %v", err)
		}
	}

	dirFS := os.DirFS(tempDir)

	b.ResetTimer() // Reset the timer to exclude setup time

	for i := 0; i < b.N; i++ {
		_ = inventory.New(dirFS, tempDir, ".distillery/bin", cfg)
	}
}

func BenchmarkInventoryHomeDir(b *testing.B) {
	userDir, _ := os.UserHomeDir()
	basePath := "/"
	baseFS := os.DirFS(basePath)
	binPath := filepath.Join(userDir, ".distillery", "bin")
	cfg, err := config.New("")
	if err != nil {
		b.Fatalf("Failed to create config: %v", err)
	}

	for i := 0; i < b.N; i++ {
		_ = inventory.New(baseFS, basePath, binPath, cfg)
	}
}
