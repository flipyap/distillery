package source

import (
	"context"
	"fmt"
	"github.com/apex/log"
	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/score"
	"github.com/google/go-github/v62/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type GitHub struct {
	Source

	client *github.Client

	Version string // Version to find for installation
	Owner   string // Owner of the repository
	Repo    string // Repository name

	Release *github.RepositoryRelease

	Assets []*GitHubAsset
}

func (s *GitHub) GetSource() string {
	return "github"
}
func (s *GitHub) GetOwner() string {
	return s.Owner
}
func (s *GitHub) GetRepo() string {
	return s.Repo
}
func (s *GitHub) GetApp() string {
	return fmt.Sprintf("%s/%s", s.Owner, s.Repo)
}

func (s *GitHub) GetDownloadsDir() string {
	return filepath.Join(s.Options.DownloadsDir, s.GetSource(), s.GetOwner(), s.GetRepo(), s.Version)
}

func (s *GitHub) GetID() string {
	return strings.Join([]string{s.GetSource(), s.GetOwner(), s.GetRepo(), s.GetOS(), s.GetArch()}, "-")
}

func (s *GitHub) Run(ctx context.Context, _, _ string) error {
	cacheFile := filepath.Join(s.Options.MetadataDir, fmt.Sprintf("cache-%s", s.GetID()))

	s.client = github.NewClient(httpcache.NewTransport(diskcache.New(cacheFile)).Client())
	githubToken := s.Options.Settings["github-token"].(string)
	if githubToken != "" {
		log.Debug("auth token provided")
		s.client = s.client.WithAuthToken(githubToken)
	}

	if err := s.FindRelease(ctx); err != nil {
		return err
	}

	if err := s.GetReleaseAssets(ctx); err != nil {
		return err
	}

	ra, err := s.FindReleaseAsset()
	if err != nil {
		return err
	}
	_ = ra

	if err := s.Download(ctx); err != nil {
		return err
	}

	defer s.Cleanup()

	if err := s.Verify(); err != nil {
		return err
	}

	if err := s.Extract(); err != nil {
		return err
	}

	if err := s.Install(); err != nil {
		return err
	}

	return nil
}

// FindRelease - query API to find the version being sought or return an error
func (s *GitHub) FindRelease(ctx context.Context) error {
	var err error
	var release *github.RepositoryRelease

	if s.Version == "latest" {
		release, _, err = s.client.Repositories.GetLatestRelease(ctx, s.GetOwner(), s.GetRepo())
		if err != nil {
			return err
		}

		s.Version = strings.TrimPrefix(release.GetTagName(), "v")
	} else {
		releases, _, err := s.client.Repositories.ListReleases(ctx, s.GetOwner(), s.GetRepo(), nil)
		if err != nil {
			return err
		}
		for _, r := range releases {
			if r.GetTagName() == s.Version || r.GetName() == fmt.Sprintf("v%s", s.Version) {
				release = r
				break
			}
		}
	}

	if release == nil {
		return fmt.Errorf("release not found")
	}

	s.Release = release

	return nil
}

func (s *GitHub) GetReleaseAssets(ctx context.Context) error {
	params := &github.ListOptions{
		PerPage: 100,
	}

	for {
		assets, res, err := s.client.Repositories.ListReleaseAssets(
			ctx, s.GetOwner(), s.GetRepo(), s.Release.GetID(), params)
		if err != nil {
			return err
		}

		for _, a := range assets {
			s.Assets = append(s.Assets, &GitHubAsset{
				Asset:        asset.New(a.GetName(), "", s.GetOS(), s.GetArch(), s.Version),
				GitHub:       s,
				ReleaseAsset: a,
			})
		}

		if res.NextPage == 0 {
			break
		}

		params.Page = res.NextPage
	}

	return nil
}

// FindReleaseAsset - find the asset that matches the current OS and Arch, if multiple matches are found it
// will attempt to find the best match based on the suffix for the appropriate OS. If no match is found an error
// is returned.
func (s *GitHub) FindReleaseAsset() (*GitHubAsset, error) {
	// 1. Setup Assets
	// 2. Determine Asset Type (checksum, archive, other, unknown)
	// 3. Score Assets
	// 4. Select best Asset Type (archive/binary)
	// 5. If Archive, we need to extract and determine which files we are keeping (binaries)
	// 6. Extract files, and copy/symlink them into place
	for _, a := range s.Assets {
		a.Score(&asset.ScoreOptions{
			OS:         s.OSConfig.GetOS(),
			Arch:       s.OSConfig.GetArchitectures(),
			Extensions: s.OSConfig.GetExtensions(),
		})

		logrus.Debugf("name: %s, score: %d", a.GetName(), a.GetScore())
	}

	var best *GitHubAsset
	for _, a := range s.Assets {
		logrus.Tracef("finding best: %s (%d)", a.GetName(), a.GetScore())
		if best == nil || a.GetScore() > best.GetScore() && (a.GetType() == asset.Archive || a.GetType() == asset.Unknown || a.GetType() == asset.Binary) {
			best = a
		}
	}

	s.Binary = best

	fileScoring := map[asset.Type][]string{}
	fileScored := map[asset.Type][]score.Sorted{}
	for _, a := range s.Assets {
		if _, ok := fileScoring[a.GetType()]; !ok {
			fileScoring[a.GetType()] = []string{}
		}
		fileScoring[a.GetType()] = append(fileScoring[a.GetType()], a.GetName())
	}
	for k, v := range fileScoring {
		var ext []string
		if k == asset.Key {
			ext = []string{"key", "pub", "pem"}
		} else if k == asset.Signature {
			ext = []string{"sig", "asc"}
		} else if k == asset.Checksum {
			ext = []string{"sha256", "md5", "sha1", "txt"}
		}

		if _, ok := fileScored[k]; !ok {
			fileScored[k] = []score.Sorted{}
		}

		fileScored[k] = score.Score(v, &score.Options{
			OS:         s.OSConfig.GetOS(),
			Arch:       s.OSConfig.GetArchitectures(),
			Extensions: ext,
			Names:      []string{strings.ReplaceAll(s.Binary.GetName(), filepath.Ext(s.Binary.GetName()), "")},
		})

		if len(fileScored[k]) > 0 {
			logrus.Debugf("file scoring sorted ! type: %d, scored: %v", k, fileScored[k][0])
		}
	}

	for _, a := range s.Assets {
		for k, v := range fileScored {
			vv := v[0]

			if a.GetType() == asset.Checksum && a.GetType() == k && a.GetName() == vv.Key {
				s.Checksum = a
			}
			if a.GetType() == asset.Signature && a.GetType() == k && a.GetName() == vv.Key {
				s.Signature = a
			}
			if a.GetType() == asset.Key && a.GetType() == k && a.GetName() == vv.Key {
				s.Key = a
			}
		}
	}

	if s.Binary != nil {
		logrus.Tracef("best binary: %s", s.Binary.GetName())
	}
	if s.Checksum != nil {
		logrus.Tracef("best checksum: %s", s.Checksum.GetName())
	}
	if s.Signature != nil {
		logrus.Tracef("best signature: %s", s.Signature.GetName())
	}
	if s.Key != nil {
		logrus.Tracef("best key: %s", s.Key.GetName())
	}

	if best != nil {
		return best, nil
	}

	return nil, fmt.Errorf("no matching asset found")
}
