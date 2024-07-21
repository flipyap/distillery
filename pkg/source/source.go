package source

import (
	"context"
	"strings"

	"github.com/ekristen/distillery/pkg/asset"
)

type ISource interface {
	GetSource() string
	GetOwner() string
	GetRepo() string
	GetApp() string
	GetID() string
	Run(context.Context, string, string) error
}

type Options struct {
	OS           string
	Arch         string
	HomeDir      string
	CacheDir     string
	BinDir       string
	MetadataDir  string
	DownloadsDir string
}

type Source struct {
	Options *Options

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

func New(source string, opts *Options) ISource {
	version := "latest"
	versionParts := strings.Split(source, "@")
	if len(versionParts) > 1 {
		source = versionParts[0]
		version = versionParts[1]
	}

	parts := strings.Split(source, "/")
	if len(parts) == 2 {
		// could be github or homebrew or hashicorp
		if parts[0] == "homebrew" {
			return &Homebrew{
				Source:  Source{Options: opts},
				Formula: parts[1],
				Version: version,
			}
		} else if parts[0] == "hashicorp" {
			return &Hashicorp{
				Source:  Source{Options: opts},
				Owner:   parts[1],
				Repo:    parts[1],
				Version: version,
			}
		}

		return &GitHub{
			Source:  Source{Options: opts},
			Owner:   parts[0],
			Repo:    parts[1],
			Version: version,
		}
	} else if len(parts) >= 3 {
		if strings.HasPrefix(parts[0], "github") {
			return &GitHub{
				Source:  Source{Options: opts},
				Owner:   parts[1],
				Repo:    parts[2],
				Version: version,
			}
		} else if strings.HasPrefix(parts[0], "gitlab") {
			return &GitLab{
				Source:  Source{Options: opts},
				Owner:   parts[1],
				Repo:    parts[2],
				Version: version,
			}
		}

		return nil
	}

	return nil
}
