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
	"github.com/ekristen/distillery/pkg/osconfig"
	"github.com/ekristen/distillery/pkg/source/homebrew"
)

type Homebrew struct {
	Source

	client *homebrew.Client

	Formula string
	Version string

	Assets []*HomebrewAsset
}

func (s *Homebrew) GetSource() string {
	return "homebrew"
}
func (s *Homebrew) GetOwner() string {
	return "homebrew"
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
	}

	detectedOS := osconfig.New(s.GetOS(), s.GetArch())

	if len(formula.Dependencies) > 0 {
		return fmt.Errorf("formula with dependencies are not currently supported")
	}

	s.Assets = make([]*HomebrewAsset, 0)
	for osSlug, variant := range formula.Bottle.Stable.Files {
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
			FileVariant: &variant,
			Homebrew:    s,
		})
	}

	for _, a := range s.Assets {
		a.Score(&asset.ScoreOptions{
			OS:         detectedOS.GetOS(),
			Arch:       detectedOS.GetArchitectures(),
			Extensions: detectedOS.GetExtensions(),
		})

		logrus.Debugf("name: %s, score: %d", a.GetName(), a.GetScore())
	}

	var best *HomebrewAsset
	for _, a := range s.Assets {
		logrus.Tracef("finding best: %s (%d)", a.GetName(), a.GetScore())
		if best == nil || a.GetScore() > best.GetScore() && (a.GetType() == asset.Archive || a.GetType() == asset.Unknown || a.GetType() == asset.Binary) {
			best = a
		}
	}

	s.Binary = best

	if best == nil {
		return fmt.Errorf("unable to find best asset")
	}

	logrus.Tracef("best found: %s", best.GetName())

	if err := best.Download(ctx); err != nil {
		return err
	}

	defer s.Cleanup()

	if err := s.Extract(); err != nil {
		return err
	}

	if err := s.Install(); err != nil {
		return err
	}

	return nil
}
