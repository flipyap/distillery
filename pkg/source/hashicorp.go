package source

import (
	"context"
	"fmt"

	"path/filepath"

	"github.com/apex/log"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"

	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/clients/hashicorp"
)

const HashicorpSource = "hashicorp"

type Hashicorp struct {
	Source

	client *hashicorp.Client

	Owner   string
	Repo    string
	Version string
}

func (s *Hashicorp) GetSource() string {
	return HashicorpSource
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
	return fmt.Sprintf("%s-%s", s.GetSource(), s.GetRepo())
}

func (s *Hashicorp) GetDownloadsDir() string {
	return filepath.Join(s.Options.DownloadsDir, s.GetSource(), s.GetOwner(), s.GetRepo(), s.Version)
}

func (s *Hashicorp) sourceRun(ctx context.Context) error {
	cacheFile := filepath.Join(s.Options.MetadataDir, fmt.Sprintf("cache-%s", s.GetID()))

	s.client = hashicorp.NewClient(httpcache.NewTransport(diskcache.New(cacheFile)).Client())

	var release *hashicorp.Release

	if s.Version == "latest" {
		releases, err := s.client.ListReleases(ctx, s.Repo, nil)
		if err != nil {
			return err
		}

		if len(releases) == 0 {
			return fmt.Errorf("no releases found for %s", s.Repo)
		}

		s.Version = releases[0].Version
		release = releases[0]
	} else {
		version, err := s.client.GetVersion(ctx, s.Repo, s.Version)
		if err != nil {
			return err
		}

		release = version
	}

	if release == nil {
		return fmt.Errorf("no release found for %s version %s", s.Repo, s.Version)
	}

	log.Infof("installing %s@%s", release.Name, release.Version)

	for _, build := range release.Builds {
		s.Assets = append(s.Assets, &HashicorpAsset{
			Asset:     asset.New(filepath.Base(build.URL), "", s.GetOS(), s.GetArch(), s.Version),
			Build:     build,
			Hashicorp: s,
			Release:   release,
		})
	}

	return nil
}

func (s *Hashicorp) Run(ctx context.Context) error {
	if err := s.sourceRun(ctx); err != nil {
		return err
	}

	if err := s.Discover(s.Assets, []string{s.Repo}); err != nil {
		return err
	}

	if err := s.commonRun(ctx); err != nil {
		return err
	}

	return nil
}
