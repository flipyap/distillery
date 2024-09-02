package source

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/clients/homebrew"
)

const HomebrewSource = "homebrew"

type Homebrew struct {
	Source

	client *homebrew.Client

	Formula string
	Version string
}

func (s *Homebrew) GetSource() string {
	return HomebrewSource
}
func (s *Homebrew) GetOwner() string {
	return HomebrewSource
}
func (s *Homebrew) GetRepo() string {
	return s.Formula
}
func (s *Homebrew) GetApp() string {
	return s.Formula
}
func (s *Homebrew) GetID() string {
	return s.Formula
}

func (s *Homebrew) GetDownloadsDir() string {
	return filepath.Join(s.Options.DownloadsDir, s.GetSource(), s.GetOwner(), s.GetRepo(), s.Version)
}

func (s *Homebrew) Run(ctx context.Context, _, _ string) error {
	cacheFile := filepath.Join(s.Options.MetadataDir, fmt.Sprintf("cache-%s", s.GetID()))

	s.client = homebrew.NewClient(httpcache.NewTransport(diskcache.New(cacheFile)).Client())

	logrus.Debug("fetching formula")

	formula, err := s.client.GetFormula(s.Formula)
	if err != nil {
		return err
	}

	if s.Version == "latest" {
		s.Version = formula.Versions.Stable
	} else {
		// match major/minor
		logrus.Debug("selecting version")
	}

	if len(formula.Dependencies) > 0 {
		return fmt.Errorf("formula with dependencies are not currently supported")
	}

	for osSlug, variant := range formula.Bottle.Stable.Files {
		newVariant := variant
		osSlug = strings.ReplaceAll(osSlug, "_", "-")
		osSlug = strings.ReplaceAll(osSlug, "x86-64", "x86_64")

		slugParts := strings.Split(osSlug, "-")
		slugArch := "amd64"
		slugCodename := slugParts[0]
		if len(slugParts) > 1 {
			slugArch = slugParts[0]
			slugCodename = slugParts[1]
		}

		name := fmt.Sprintf("%s-%s-%s-%s", formula.Name, s.Version, slugCodename, slugArch)

		s.Assets = append(s.Assets, &HomebrewAsset{
			Asset:       asset.New(name, "", s.GetOS(), s.GetArch(), s.Version),
			FileVariant: &newVariant,
			Homebrew:    s,
		})
	}

	if err := s.Discover(s.Assets, []string{s.Formula}); err != nil {
		return err
	}

	if err := s.Download(ctx); err != nil {
		return err
	}

	defer func(s *Homebrew) {
		_ = s.Cleanup()
	}(s)

	if err := s.Extract(); err != nil {
		return err
	}

	if err := s.Install(); err != nil {
		return err
	}

	return nil
}
