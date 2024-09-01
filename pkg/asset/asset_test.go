package asset

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAsset(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		expectType  Type
		expectScore int
	}{
		{"test", "Test", Unknown, 0},
		{"test.tar.gz", "Test", Archive, 0},
		{"test.tar.gz.asc", "Test", Signature, 0},
		{"dist.tar.gz.sig", "dist.tar.gz.sig", Signature, 0},
	}

	for _, c := range cases {
		asset := New(c.name, c.displayName, "linux", "amd64", "1.0.0")

		if asset.GetName() != c.name {
			t.Errorf("expected name to be %s, got %s", c.name, asset.GetName())
		}
		if asset.GetDisplayName() != c.displayName {
			t.Errorf("expected display name to be %s, got %s", c.displayName, asset.GetDisplayName())
		}
		if asset.Type != c.expectType {
			t.Errorf("expected type to be %d, got %d", c.expectType, asset.Type)
		}
		if asset.score != c.expectScore {
			t.Errorf("expected score to be %d, got %d", c.expectScore, asset.score)
		}
	}
}

func TestAssetScoring(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		scoringOpts *ScoreOptions
		expectType  Type
		expectScore int
	}{
		{"test.tar.gz", "Test", &ScoreOptions{OS: []string{"linux"}, Arch: []string{"amd64"}, Extensions: []string{".tar.gz"}}, Archive, 15},
		{"test_amd64.tar.gz", "Test", &ScoreOptions{OS: []string{"linux"}, Arch: []string{"amd64"}, Extensions: []string{".tar.gz"}}, Archive, 20},
		{"test_linux_amd64.tar.gz", "Test", &ScoreOptions{OS: []string{"linux"}, Arch: []string{"amd64"}, Extensions: []string{"unknown", ".tar.gz"}}, Archive, 30},
		{"test_linux_x86_64.tar.gz", "Test", &ScoreOptions{OS: []string{"Linux"}, Arch: []string{"amd64", "x86_64"}, Extensions: []string{"unknown", ".tar.gz"}}, Archive, 30},
	}

	for _, c := range cases {
		asset := New(c.name, c.displayName, "linux", "amd64", "1.0.0")
		asset.Score(c.scoringOpts)

		if asset.GetName() != c.name {
			t.Errorf("expected name to be %s, got %s", c.name, asset.GetName())
		}
		if asset.GetDisplayName() != c.displayName {
			t.Errorf("expected display name to be %s, got %s", c.displayName, asset.GetDisplayName())
		}
		if asset.Type != c.expectType {
			t.Errorf("expected type to be %d, got %d", c.expectType, asset.Type)
		}
		if asset.score != c.expectScore {
			t.Errorf("expected score to be %d, got %d", c.expectScore, asset.score)
		}
	}
}

func TestDefaultAsset(t *testing.T) {
	asset := New("dist-linux-amd64.tar.gz", "dist-linux-amd64.tar.gz", "linux", "amd64", "1.0.0")
	asset.Score(&ScoreOptions{OS: []string{"linux"}, Arch: []string{"amd64"}, Extensions: []string{".tar.gz"}})
	err := asset.Download(context.TODO())
	assert.Error(t, err)

	assert.Equal(t, Archive, asset.GetType())
	assert.Equal(t, "not-implemented", asset.ID())
	assert.Equal(t, 30, asset.GetScore())
	assert.Equal(t, "dist-linux-amd64.tar.gz", asset.GetDisplayName())
	assert.Equal(t, "dist-linux-amd64.tar.gz", asset.GetName())
	assert.Equal(t, "", asset.GetFilePath())
	assert.Equal(t, "", asset.GetTempPath())
	assert.Equal(t, asset, asset.GetAsset())
	assert.Equal(t, make([]*File, 0), asset.GetFiles())
}

func TestAssetTypes(t *testing.T) {
	cases := []struct {
		name     string
		fileType Type
	}{
		{
			name:     "dist-linux-amd64.deb",
			fileType: Installer,
		},
		{
			name:     "dist-linux-amd64.rpm",
			fileType: Installer,
		},
		{
			name:     "dist-linux-amd64.tar.gz",
			fileType: Archive,
		},
		{
			name:     "dist-linux-amd64.exe",
			fileType: Binary,
		},
		{
			name:     "dist-linux-amd64",
			fileType: Unknown,
		},
		{
			name:     "dist-linux-amd64.tar.gz.sig",
			fileType: Signature,
		},
		{
			name:     "dist-linux-amd64.tar.gz.pem",
			fileType: Key,
		},
		{
			name:     "checksums.txt",
			fileType: Checksum,
		},
		{
			name:     "dist-linux.SHASUMS",
			fileType: Checksum,
		},
		{
			name:     "dist-linux-amd64.tar.gz.sha256",
			fileType: Checksum,
		},
		{
			name:     "dist-linux.nse",
			fileType: Unknown,
		},
		{
			name:     "dist-linux.deb",
			fileType: Installer,
		},
		{
			name:     "dist-windows.msi",
			fileType: Installer,
		},
	}

	for _, c := range cases {
		asset := New(c.name, c.name, "linux", "amd64", "1.0.0")
		assert.Equal(t, c.fileType, asset.GetType(), fmt.Sprintf("expected type to be %d, got %d for %s", c.fileType, asset.GetType(), c.name))
	}
}
