package checksum

import (
	"bufio"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func ComputeFileHash(filePath string, hashFunc func() hash.Hash) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := hashFunc()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// CompareHashWithChecksumFile compares the computed hash of a file with the hashes in a checksum file.
func CompareHashWithChecksumFile(fileName, filePath, checksumFilePath string, hashFunc func() hash.Hash) (bool, error) {
	log := logrus.WithField("handler", "compare-hash-with-checksum-file")

	// Compute the hash of the file
	computedHash, err := ComputeFileHash(filePath, hashFunc)
	if err != nil {
		return false, err
	}

	// Open the checksum file
	checksumFile, err := os.Open(checksumFilePath)
	if err != nil {
		return false, err
	}
	defer checksumFile.Close()

	// Read and compare hashes
	scanner := bufio.NewScanner(checksumFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			log.Trace("skipping line: ", line)
			continue
		}
		fileHash := parts[0]
		filename := parts[1]
		log.Trace("fileHash: ", fileHash)
		log.Trace("filename: ", filename)
		// Rust does *(binary) for the binary name
		filename = strings.TrimPrefix(filename, "*")

		if (filename == fileName || filepath.Base(filename) == fileName) && fileHash == computedHash {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}
