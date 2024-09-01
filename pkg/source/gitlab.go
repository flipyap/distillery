package source

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/osconfig"
	"github.com/ekristen/distillery/pkg/source/gitlab"
)

type GitLab struct {
	Source

	client *gitlab.Client

	Owner   string
	Repo    string
	Version string

	Release *gitlab.Release

	Assets []*GitLabAsset
}

func (s *GitLab) GetSource() string {
	return "gitlab"
}
func (s *GitLab) GetOwner() string {
	return s.Owner
}
func (s *GitLab) GetRepo() string {
	return s.Repo
}
func (s *GitLab) GetApp() string {
	return fmt.Sprintf("%s/%s", s.Owner, s.Repo)
}
func (s *GitLab) GetID() string {
	return fmt.Sprintf("%s/%s/%s", s.GetSource(), s.GetOwner(), s.GetRepo())
}

func (s *GitLab) GetDownloadsDir() string {
	return filepath.Join(s.Options.DownloadsDir, s.GetSource(), s.GetOwner(), s.GetRepo(), s.Version)
}

func (s *GitLab) Run(ctx context.Context, _, _ string) error {
	cacheFile := filepath.Join(s.Options.MetadataDir, fmt.Sprintf("cache-%s", s.GetID()))

	s.client = gitlab.NewClient(httpcache.NewTransport(diskcache.New(cacheFile)).Client())
	token := s.Options.Settings["gitlab-token"].(string)
	if token != "" {
		s.client.SetToken(token)
	}

	if s.Version == "latest" {
		release, err := s.client.GetLatestRelease(fmt.Sprintf("%s/%s", s.Owner, s.Repo))
		if err != nil {
			return err
		}

		s.Version = release.TagName
		s.Release = release
	} else {
		release, err := s.client.GetRelease(fmt.Sprintf("%s/%s", s.Owner, s.Repo), s.Version)
		if err != nil {
			return err
		}

		s.Release = release
	}

	if s.Release == nil {
		return fmt.Errorf("no release found for %s version %s", s.GetApp(), s.Version)
	}

	for _, a := range s.Release.Assets.Links {
		s.Assets = append(s.Assets, &GitLabAsset{
			Asset:  asset.New(filepath.Base(a.URL), "", s.GetOS(), s.GetArch(), s.Version),
			GitLab: s,
			Link:   a,
		})
	}

	detectedOS := osconfig.New(s.GetOS(), s.GetArch())

	for _, a := range s.Assets {
		a.Score(&asset.ScoreOptions{
			OS:         detectedOS.GetOS(),
			Arch:       detectedOS.GetArchitectures(),
			Extensions: detectedOS.GetExtensions(),
		})

		logrus.Debugf("name: %s, score: %d", a.GetName(), a.GetScore())
	}

	var best *GitLabAsset
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

	logrus.Tracef("best: %s", best.GetName())

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
