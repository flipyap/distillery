package source

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ekristen/distillery/pkg/cosign"
	"github.com/sirupsen/logrus"
	"os"
	"strings"

	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/osconfig"
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
	logrus.Info("download called")
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
	/*
		if err := s.verifySignature(); err != nil {
			return err
		}
	*/

	return s.verifyChecksum()
}

func (s *Source) verifySignature() error {
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

	version := "latest"
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
		if parts[0] == "homebrew" {
			return &Homebrew{
				Source:  Source{Options: opts, OSConfig: detectedOS},
				Formula: parts[1],
				Version: version,
			}, nil
		} else if parts[0] == "hashicorp" {
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
			if parts[1] == "hashicorp" {
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
