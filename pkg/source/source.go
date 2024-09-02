package source

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/checksum"
	"github.com/ekristen/distillery/pkg/cosign"
	"github.com/ekristen/distillery/pkg/osconfig"
)

const (
	VersionLatest = "latest"
)

type ISource interface {
	GetSource() string
	GetOwner() string
	GetRepo() string
	GetApp() string
	GetID() string
	GetDownloadsDir() string
	Run(context.Context, string, string) error
}

type Options struct {
	OS           string
	Arch         string
	HomeDir      string
	CacheDir     string
	BinDir       string
	OptDir       string
	MetadataDir  string
	DownloadsDir string

	Settings map[string]interface{}
}

type Source struct {
	Options  *Options
	OSConfig *osconfig.OS

	File string

	Binary    asset.IAsset
	Signature asset.IAsset
	Checksum  asset.IAsset
	Key       asset.IAsset
}

func (s *Source) GetOS() string {
	return s.Options.OS
}

func (s *Source) GetArch() string {
	return s.Options.Arch
}

func (s *Source) Download(ctx context.Context) error {
	log.Info("downloading assets")
	if s.Binary != nil {
		if err := s.Binary.Download(ctx); err != nil {
			return err
		}
	}

	if s.Signature != nil {
		if err := s.Signature.Download(ctx); err != nil {
			return err
		}
	}

	if s.Checksum != nil {
		if err := s.Checksum.Download(ctx); err != nil {
			return err
		}
	}

	if s.Key != nil {
		if err := s.Key.Download(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Source) Verify() error {
	if err := s.verifyChecksum(); err != nil {
		return err
	}

	return s.verifySignature()
}

func (s *Source) verifySignature() error {
	if true {
		logrus.Debug("skipping signature verification")
		return nil
	}

	logrus.Info("verifying signature")

	cosignFileContent, err := os.ReadFile(s.Checksum.GetFilePath())
	if err != nil {
		return err
	}

	publicKeyContentEncoded, err := os.ReadFile(s.Key.GetFilePath())
	if err != nil {
		return err
	}

	publicKeyContent, err := base64.StdEncoding.DecodeString(string(publicKeyContentEncoded))
	if err != nil {
		return err
	}

	pubKey, err := cosign.ParsePublicKey(publicKeyContent)
	if err != nil {
		return err
	}

	fmt.Printf("Public Key: %+v\n", pubKey)

	sigData, err := os.ReadFile(s.Signature.GetFilePath())
	if err != nil {
		return err
	}

	valid, err := cosign.VerifySignature(pubKey, cosignFileContent, sigData)
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("unable to validate signature")
	}

	return nil
}

func (s *Source) verifyChecksum() error {
	if v, ok := s.Options.Settings["no-checksum-verify"]; ok && v.(bool) {
		log.Warn("skipping checksum verification")
		return nil
	}

	if s.Checksum == nil {
		log.Warn("skipping checksum verification (no checksum)")
		return nil
	}

	logrus.Debug("verifying checksum")
	logrus.Tracef("binary: %s", s.Binary.GetName())

	match, err := checksum.CompareHashWithChecksumFile(s.Binary.GetName(),
		s.Binary.GetFilePath(), s.Checksum.GetFilePath(), sha256.New)
	if err != nil {
		return err
	}

	logrus.Tracef("checksum match: %v", match)

	if !match {
		return fmt.Errorf("checksum verification failed")
	}

	log.Info("checksum verified")

	return nil
}

func (s *Source) Extract() error {
	return s.Binary.Extract()
}

func (s *Source) Install() error {
	return s.Binary.Install(s.Binary.ID(), s.Options.BinDir)
}

func (s *Source) Cleanup() error {
	return s.Binary.Cleanup()
}

func New(source string, opts *Options) (ISource, error) {
	detectedOS := osconfig.New(opts.OS, opts.Arch)

	version := VersionLatest
	versionParts := strings.Split(source, "@")
	if len(versionParts) > 1 {
		source = versionParts[0]
		version = versionParts[1]
	}

	parts := strings.Split(source, "/")

	if len(parts) == 1 {
		return nil, fmt.Errorf("invalid install source, expect format of owner/repo or owner/repo@version")
	}

	if len(parts) == 2 {
		// could be GitHub or Homebrew or Hashicorp
		if parts[0] == HomebrewSource {
			return &Homebrew{
				Source:  Source{Options: opts, OSConfig: detectedOS},
				Formula: parts[1],
				Version: version,
			}, nil
		} else if parts[0] == HashicorpSource {
			return &Hashicorp{
				Source:  Source{Options: opts, OSConfig: detectedOS},
				Owner:   parts[1],
				Repo:    parts[1],
				Version: version,
			}, nil
		}

		return &GitHub{
			Source:  Source{Options: opts, OSConfig: detectedOS},
			Owner:   parts[0],
			Repo:    parts[1],
			Version: version,
		}, nil
	} else if len(parts) >= 3 {
		if strings.HasPrefix(parts[0], "github") {
			if parts[1] == HashicorpSource {
				return &Hashicorp{
					Source:  Source{Options: opts, OSConfig: detectedOS},
					Owner:   parts[1],
					Repo:    parts[2],
					Version: version,
				}, nil
			}

			return &GitHub{
				Source:  Source{Options: opts, OSConfig: detectedOS},
				Owner:   parts[1],
				Repo:    parts[2],
				Version: version,
			}, nil
		} else if strings.HasPrefix(parts[0], "gitlab") {
			return &GitLab{
				Source:  Source{Options: opts, OSConfig: detectedOS},
				Owner:   parts[1],
				Repo:    parts[2],
				Version: version,
			}, nil
		}

		return nil, nil
	}

	return nil, nil
}
