package asset

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"github.com/krolaw/zipstream"
	"github.com/sirupsen/logrus"
	"github.com/xi2/xz"

	"github.com/ekristen/distillery/pkg/common"
)

var (
	msiType      = filetype.AddType("msi", "application/octet-stream")
	apkType      = filetype.AddType("apk", "application/vnd.android.package-archive")
	ascType      = filetype.AddType("asc", "text/plain")
	pemType      = filetype.AddType("pem", "application/x-pem-file")
	sigType      = filetype.AddType("sig", "text/plain")
	sbomJSONType = filetype.AddType("sbom.json", "application/json")
	bomJSONType  = filetype.AddType("bom.json", "application/json")
	jsonType     = filetype.AddType("json", "application/json")
	sbomType     = filetype.AddType("sbom", "application/octet-stream")
	bomType      = filetype.AddType("bom", "application/octet-stream")
	pubType      = filetype.AddType("pub", "text/plain")
	tarGzType    = filetype.AddType("tgz", "application/tar+gzip")

	ignoreFileExtensions = []string{
		".txt",
		".sbom",
		".json",
	}

	executableMimetypes = []string{
		"application/x-mach-binary",
		"application/x-executable",
		"application/x-elf",
		"application/vnd.microsoft.portable-executable",
	}
)

// Type is the type of asset
type Type int

const (
	Unknown Type = iota
	Archive
	Binary
	Installer
	Checksum
	Signature
	Key
	SBOM
	Data
)

// processorFunc is a function that processes a reader
type processorFunc func(io.Reader) (io.Reader, error)

// New creates a new asset
func New(name, displayName, osName, osArch, version string) *Asset {
	a := &Asset{
		Name:        name,
		DisplayName: displayName,
		OS:          osName,
		Arch:        osArch,
		Version:     version,
		Files:       make([]*File, 0),
	}

	a.Classify()

	return a
}

type File struct {
	Name        string
	Alias       string
	Installable bool
}

type Asset struct {
	Name        string
	DisplayName string
	Type        Type

	OS      string
	Arch    string
	Version string

	Extension    string
	DownloadPath string
	Hash         string
	TempDir      string
	Files        []*File
}

func (a *Asset) ID() string {
	return "not-implemented"
}

func (a *Asset) GetName() string {
	return a.Name
}

func (a *Asset) GetDisplayName() string {
	return a.DisplayName
}

func (a *Asset) GetType() Type {
	return a.Type
}

func (a *Asset) GetAsset() *Asset {
	return a
}

func (a *Asset) GetFiles() []*File {
	return a.Files
}
func (a *Asset) GetTempPath() string {
	return a.TempDir
}

func (a *Asset) Download(_ context.Context) error {
	return fmt.Errorf("not implemented")
}

func (a *Asset) GetFilePath() string {
	return a.DownloadPath
}

// Classify determines the type of asset based on the file extension
func (a *Asset) Classify() { //nolint:gocyclo
	if ext := strings.TrimPrefix(filepath.Ext(a.Name), "."); ext != "" {
		switch filetype.GetType(ext) {
		case matchers.TypeDeb, matchers.TypeRpm, msiType, apkType:
			a.Type = Installer
		case matchers.TypeGz, matchers.TypeZip, matchers.TypeXz, matchers.TypeTar, matchers.TypeBz2, tarGzType:
			a.Type = Archive
		case matchers.TypeExe:
			a.Type = Binary
		case sigType, ascType:
			a.Type = Signature
		case pemType, pubType:
			a.Type = Key
		case sbomJSONType, bomJSONType, sbomType, bomType:
			a.Type = SBOM
		case jsonType:
			a.Type = Data

			if strings.Contains(a.Name, ".sbom") || strings.Contains(a.Name, ".bom") {
				a.Type = SBOM
			}
		default:
			a.Type = Unknown
		}
	}

	if a.Type == Unknown {
		logrus.Tracef("classifying asset based on name: %s", a.Name)
		name := strings.ToLower(a.Name)
		if strings.Contains(name, ".sha256") || strings.Contains(name, ".md5") {
			a.Type = Checksum
		}
		if strings.Contains(name, "checksums") {
			a.Type = Checksum
		}
		if strings.Contains(a.Name, "SHA") && strings.Contains(a.Name, "SUMS") {
			a.Type = Checksum
		} else if strings.Contains(a.Name, "SUMS") {
			a.Type = Checksum
		}

		if strings.Contains(a.Name, "-pivkey-") {
			a.Type = Key
		} else if strings.Contains(a.Name, "pkcs") && strings.Contains(a.Name, "key") {
			a.Type = Key
		}
	}

	logrus.Tracef("classified: %s (%d)", a.Name, a.Type)
}

func (a *Asset) copyFile(srcFile, dstFile string) error {
	// Open the source file for reading
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstFile, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	return nil
}

// Install installs the asset
// TODO(ek): simplify this function
func (a *Asset) Install(id, binDir string) error { //nolint:funlen
	found := false

	logrus.Tracef("files to process: %d", len(a.Files))
	for _, file := range a.Files {
		// Actual path to the downloaded/extracted file
		fullPath := filepath.Join(a.TempDir, file.Name)

		logrus.Debug("checking file for installable: ", file.Name)
		m, err := mimetype.DetectFile(fullPath)
		if err != nil {
			return err
		}

		logrus.Debug("found mimetype: ", m.String())

		if slices.Contains(ignoreFileExtensions, m.Extension()) {
			logrus.Tracef("ignoring file: %s", file.Name)
			continue
		}

		if slices.Contains(executableMimetypes, m.String()) {
			logrus.Debugf("found installable executable: %s, %s, %s", file.Name, m.String(), m.Extension())
			file.Installable = true
		}
	}

	for _, file := range a.Files {
		if !file.Installable {
			logrus.Tracef("skipping file: %s", file.Name)
			continue
		}

		found = true
		logrus.Debugf("installing file: %s", file.Name)

		fullPath := filepath.Join(a.TempDir, file.Name)
		dstFilename := filepath.Base(fullPath)
		if file.Alias != "" {
			dstFilename = file.Alias
		}

		logrus.Trace("pre-dstFilename: ", dstFilename)

		// Strip the OS and Arch from the filename if it exists, this happens mostly when the binary is being
		// uploaded directly instead of being encapsulated in a tarball or zip file
		dstFilename = strings.ReplaceAll(dstFilename, a.OS, "")
		dstFilename = strings.ReplaceAll(dstFilename, a.Arch, "")

		dstFilename = strings.ReplaceAll(dstFilename, fmt.Sprintf("v%s", a.Version), "")
		dstFilename = strings.ReplaceAll(dstFilename, a.Version, "")

		dstFilename = strings.TrimSpace(dstFilename)
		dstFilename = strings.TrimRight(dstFilename, "-")
		dstFilename = strings.TrimRight(dstFilename, "_")

		logrus.Tracef("post-dstFilename: %s", dstFilename)

		destBinaryName := fmt.Sprintf("%s-%s", id, dstFilename)
		destBinFilename := filepath.Join(binDir, destBinaryName)
		defaultBinFilename := filepath.Join(binDir, dstFilename)

		versionedBinFilename := fmt.Sprintf("%s@%s", defaultBinFilename, strings.TrimLeft(a.Version, "v"))

		logrus.Debugf("copying executable: %s to %s", fullPath, destBinFilename)
		if err := a.copyFile(fullPath, destBinFilename); err != nil {
			return err
		}

		// create symlink
		// TODO: check if symlink exists
		// TODO: allow override
		if runtime.GOOS == a.OS && runtime.GOARCH == a.Arch {
			logrus.Debugf("creating symlink: %s to %s", defaultBinFilename, destBinFilename)
			logrus.Debugf("creating symlink: %s to %s", versionedBinFilename, destBinFilename)
			_ = os.Remove(defaultBinFilename)
			_ = os.Remove(versionedBinFilename)
			_ = os.Symlink(destBinFilename, defaultBinFilename)
			_ = os.Symlink(destBinFilename, versionedBinFilename)
		}
	}

	if !found {
		return fmt.Errorf("the request binary was not found in the release")
	}

	return nil
}

func (a *Asset) Cleanup() error {
	logrus.WithField("asset", a.GetName()).Tracef("cleaning up temp dir: %s", a.TempDir)
	return os.RemoveAll(a.TempDir)
}

func (a *Asset) Extract() error {
	var err error

	fileHandler, err := os.Open(a.DownloadPath)
	if err != nil {
		return err
	}

	a.TempDir, err = os.MkdirTemp("", common.NAME)
	if err != nil {
		return err
	}

	logrus.Debugf("opened and extracting file: %s", a.DownloadPath)

	return a.doExtract(fileHandler)
}

func (a *Asset) doExtract(in io.Reader) error {
	var buf bytes.Buffer
	tee := io.TeeReader(in, &buf)

	t, err := filetype.MatchReader(tee)
	if err != nil {
		return err
	}

	outputFile := io.MultiReader(&buf, in)

	logrus.Debugf("extracting file type: %s", t)

	var processor processorFunc

	switch t {
	case matchers.TypeTar:
		processor = a.processTar
	case matchers.TypeZip:
		processor = a.processZip
	case matchers.TypeBz2:
		processor = a.processBz2
	case matchers.TypeGz:
		processor = a.processGz
	case matchers.TypeXz:
		processor = a.processXz
	default:
		processor = a.processDirect
	}

	if processor != nil {
		newReader, err := processor(outputFile)
		if err != nil {
			return err
		}

		if newReader == nil {
			return nil
		}

		// In case of e.g. a .tar.gz, process the uncompressed archive by calling recursively
		return a.doExtract(newReader)
	}

	return nil
}

func (a *Asset) processDirect(in io.Reader) (io.Reader, error) {
	logrus.Tracef("processing direct file")
	outFile, err := os.Create(filepath.Join(a.TempDir, filepath.Base(a.DownloadPath)))
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(outFile, in); err != nil {
		return nil, err
	}

	a.Files = append(a.Files, &File{Name: filepath.Base(a.DownloadPath), Alias: a.GetName()})

	return nil, nil
}

func (a *Asset) processZip(in io.Reader) (io.Reader, error) {
	zr := zipstream.NewReader(in)
	a.Files = make([]*File, 0)

	for {
		header, err := zr.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		target := filepath.Join(a.TempDir, header.Name)
		logrus.Tracef("zip > target %s", target)

		if header.Mode().IsDir() {
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return nil, err
				}
				logrus.Tracef("tar > create directory %s", target)
			}

			continue
		}

		// TODO(ek): do we need to somehow check the location in the zip file?
		// TODO(ek): should we cache the hashes of the files back to the main hash of the file?

		f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, header.Mode())
		if err != nil {
			return nil, err
		}

		// copy over contents
		if _, err := io.Copy(f, zr); err != nil {
			return nil, err
		}

		// manually close here after each file operation; deferring would cause each file close
		// to wait until all operations have completed.
		f.Close()

		a.Files = append(a.Files, &File{Name: header.Name})
		logrus.Tracef("zip > create file %s", target)
	}

	if len(a.Files) == 0 {
		return nil, fmt.Errorf("no files found in zip archive")
	}

	return nil, nil
}

func (a *Asset) processTar(in io.Reader) (io.Reader, error) {
	logrus.Trace("processing tar file")
	tr := tar.NewReader(in)
	a.Files = make([]*File, 0)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// TODO(ek): do we need to somehow check the location in the tar file?

		target := filepath.Join(a.TempDir, header.Name) //nolint:gosec
		logrus.Tracef("tar > target %s", target)

		switch header.Typeflag {
		// if it's a dir, and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return nil, err
				}
				logrus.Tracef("tar > create directory %s", target)
			}
		// if it's a file create it
		case tar.TypeReg:
			baseDir := filepath.Dir(target)
			if _, err := os.Stat(baseDir); err != nil {
				if err := os.MkdirAll(baseDir, 0755); err != nil {
					return nil, err
				}
				logrus.Tracef("tar > create directory %s", baseDir)
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return nil, err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil { //nolint: gosec
				return nil, err
			}

			// manually close here after each file operation; deferring would cause each file close
			// to wait until all operations have completed.
			f.Close()

			a.Files = append(a.Files, &File{Name: header.Name})
			logrus.Tracef("tar > create file %s", target)
		}
	}

	if len(a.Files) == 0 {
		return nil, fmt.Errorf("no files in tar archive")
	}

	return nil, nil
}

func (a *Asset) processGz(in io.Reader) (io.Reader, error) {
	gr, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}

	return gr, nil
}

func (a *Asset) processXz(in io.Reader) (io.Reader, error) {
	xr, err := xz.NewReader(in, 0)
	if err != nil {
		return nil, err
	}

	return xr, nil
}

func (a *Asset) processBz2(in io.Reader) (io.Reader, error) {
	br := bzip2.NewReader(in)
	return br, nil
}
