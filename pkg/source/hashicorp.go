package source //nolint:dupl

import (
	"context"
	"fmt"
	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/osconfig"
	"github.com/ekristen/distillery/pkg/source/hashicorp"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

type Hashicorp struct {
	Source

	client *hashicorp.Client

	Owner   string
	Repo    string
	Version string

	Assets []*HashicorpAsset
}

func (s *Hashicorp) GetSource() string {
	return "hashicorp"
}
func (s *Hashicorp) GetOwner() string {
	return s.Owner
}
func (s *Hashicorp) GetRepo() string {
	return s.Repo
}
func (s *Hashicorp) GetApp() string {
	return fmt.Sprintf("%s/%s", s.Owner, s.Repo)
}
func (s *Hashicorp) GetID() string {
	return fmt.Sprintf("%s/%s/%s", s.GetSource(), s.GetOwner(), s.GetRepo())
}
func (s *Hashicorp) Run(ctx context.Context, _, _ string) error {
	cacheFile := filepath.Join(s.Options.MetadataDir, fmt.Sprintf("cache-%s", s.GetID()))

	s.client = hashicorp.NewClient(httpcache.NewTransport(diskcache.New(cacheFile)).Client())

	var release *hashicorp.Release

	if s.Version == "latest" {
		releases, err := s.client.ListReleases(s.Repo)
		if err != nil {
			return err
		}

		if len(releases) == 0 {
			return fmt.Errorf("no releases found for %s", s.Repo)
		}

		s.Version = releases[0].Version
		release = releases[0]
	} else {
		version, err := s.client.GetVersion(s.Repo, s.Version)
		if err != nil {
			return err
		}

		release = version
	}

	if release == nil {
		return fmt.Errorf("no release found for %s version %s", s.Repo, s.Version)
	}

	detectedOS := osconfig.New(s.GetOS(), s.GetArch())

	fmt.Println(release.Name, release.Version)

	s.Assets = make([]*HashicorpAsset, 0)
	for _, build := range release.Builds {
		s.Assets = append(s.Assets, &HashicorpAsset{
			Asset:     asset.New(filepath.Base(build.URL), "", s.GetOS(), s.GetArch(), s.Version),
			Build:     build,
			Hashicorp: s,
			Release:   release,
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

	var best *HashicorpAsset
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
