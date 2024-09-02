package asset

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dsnet/compress/bzip2"
	"github.com/stretchr/testify/assert"
	"github.com/ulikunitz/xz"
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
		{
			"test.tar.gz",
			"Test",
			&ScoreOptions{OS: []string{"linux"}, Arch: []string{"amd64"}, Extensions: []string{".tar.gz"}},
			Archive,
			15,
		},
		{
			"test_amd64.tar.gz",
			"Test",
			&ScoreOptions{OS: []string{"linux"}, Arch: []string{"amd64"}, Extensions: []string{".tar.gz"}},
			Archive,
			20,
		},
		{
			"test_linux_amd64.tar.gz",
			"Test",
			&ScoreOptions{OS: []string{"linux"}, Arch: []string{"amd64"}, Extensions: []string{"unknown", ".tar.gz"}},
			Archive,
			30,
		},
		{
			"test_linux_x86_64.tar.gz",
			"Test",
			&ScoreOptions{
				OS: []string{"Linux"}, Arch: []string{"amd64", "x86_64"}, Extensions: []string{"unknown", ".tar.gz"},
			},
			Archive,
			30,
		},
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
		{
			name:     "dist-linux-amd64.sbom.json",
			fileType: SBOM,
		},
		{
			name:     "dist-linux-amd64.json",
			fileType: Unknown,
		},
		{
			name:     "dist-linux-amd64.sbom",
			fileType: SBOM,
		},
	}

	for _, c := range cases {
		asset := New(c.name, c.name, "linux", "amd64", "1.0.0")
		assert.Equal(t, c.fileType, asset.GetType(), fmt.Sprintf("expected type to be %d, got %d for %s", c.fileType, asset.GetType(), c.name))
	}
}

func TestAssetExtract(t *testing.T) {
	cases := []struct {
		name         string
		fileType     Type
		internalFile string
		downloadFile string
	}{
		{
			name:         "dist-linux-amd64.tar.gz",
			fileType:     Archive,
			internalFile: "test-file",
			downloadFile: createTarGz(t, "test-file", "This is a test file content"),
		},
		{
			name:         "dist-linux-amd64.zip",
			fileType:     Archive,
			internalFile: "test-file",
			downloadFile: createZip(t, "test-file", "This is a test file content"),
		},
		{
			name:         "dist-linux-amd64.tar.bz2",
			fileType:     Archive,
			internalFile: "test-file",
			downloadFile: createTarBz2(t, "test-file", "This is a test file content"),
		},
		{
			name:         "dist-linux-amd64.tar.xz",
			fileType:     Archive,
			internalFile: "test-file",
			downloadFile: createTarXz(t, "test-file", "This is a test file content"),
		},
		{
			name:         "dist-linux-amd64",
			fileType:     Binary,
			internalFile: "test-*",
			downloadFile: createFile(t, "This is a test file content"),
		},
		{
			name:         "windows-executable",
			fileType:     Binary,
			internalFile: "test-*",
			downloadFile: createFile(t, "This is a test file content"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			asset := New(c.name, c.name, "linux", "amd64", "1.0.0")
			asset.DownloadPath = c.downloadFile

			defer func(asset *Asset) {
				_ = asset.Cleanup()
			}(asset)

			defer func(path string) {
				_ = os.RemoveAll(path)
			}(c.downloadFile)

			err := asset.Extract()
			assert.NoError(t, err)

			// Verify the asset
			if strings.HasSuffix(c.internalFile, "-*") {
				assert.True(t, strings.HasPrefix(asset.Files[0].Name, "test-"))
			} else {
				assert.Equal(t, c.internalFile, asset.Files[0].Name)
			}

			assert.Equal(t, 1, len(asset.Files))
		})
	}
}

func createFile(t *testing.T, content string) string {
	t.Helper()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*")
	assert.NoError(t, err)
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)

	return tmpFile.Name()
}

func createTarGz(t *testing.T, fileName, content string) string {
	t.Helper()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.tar.gz")
	assert.NoError(t, err)
	defer tmpFile.Close()

	// Create a gzip writer
	gw := gzip.NewWriter(tmpFile)
	defer gw.Close()

	// Create a tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add a file to the tar archive
	hdr := &tar.Header{
		Name: fileName,
		Mode: 0600,
		Size: int64(len(content)),
	}
	err = tw.WriteHeader(hdr)
	assert.NoError(t, err)

	_, err = tw.Write([]byte(content))
	assert.NoError(t, err)

	return tmpFile.Name()
}

func createTarBz2(t *testing.T, fileName, content string) string {
	t.Helper()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.tar.bz2")
	assert.NoError(t, err)
	defer tmpFile.Close()

	// Create a bzip2 writer
	bw, err := bzip2.NewWriter(tmpFile, &bzip2.WriterConfig{Level: bzip2.BestCompression})
	assert.NoError(t, err)
	defer bw.Close()

	// Create a tar writer
	tw := tar.NewWriter(bw)
	defer tw.Close()

	// Add a file to the tar archive
	hdr := &tar.Header{
		Name: fileName,
		Mode: 0600,
		Size: int64(len(content)),
	}
	err = tw.WriteHeader(hdr)
	assert.NoError(t, err)

	_, err = tw.Write([]byte(content))
	assert.NoError(t, err)

	return tmpFile.Name()
}

func createTarXz(t *testing.T, fileName, content string) string {
	t.Helper()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.tar.xz")
	assert.NoError(t, err)
	defer tmpFile.Close()

	// Create an xz writer
	xw, err := xz.NewWriter(tmpFile)
	assert.NoError(t, err)
	defer xw.Close()

	// Create a tar writer
	tw := tar.NewWriter(xw)
	defer tw.Close()

	// Add a file to the tar archive
	hdr := &tar.Header{
		Name: fileName,
		Mode: 0600,
		Size: int64(len(content)),
	}
	err = tw.WriteHeader(hdr)
	assert.NoError(t, err)

	_, err = tw.Write([]byte(content))
	assert.NoError(t, err)

	return tmpFile.Name()
}

func createZip(t *testing.T, fileName, content string) string {
	t.Helper()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.zip")
	assert.NoError(t, err)
	defer tmpFile.Close()

	// Create a zip writer
	zw := zip.NewWriter(tmpFile)
	defer zw.Close()

	// Add a file to the zip archive
	w, err := zw.Create(fileName)
	assert.NoError(t, err)

	_, err = io.Copy(w, bytes.NewReader([]byte(content)))
	assert.NoError(t, err)

	return tmpFile.Name()
}
