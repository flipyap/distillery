package source

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/v62/github"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/common"
)

type GitHubAsset struct {
	*asset.Asset

	GitHub       *GitHub
	ReleaseAsset *github.ReleaseAsset
}

func (a *GitHubAsset) ID() string {
	return fmt.Sprintf("%s-%s-%s-%d", a.GitHub.GetOwner(), a.GitHub.GetRepo(), a.GitHub.Version, a.ReleaseAsset.GetID())
}

func (a *GitHubAsset) Download(ctx context.Context) error {
	rc, url, err := a.GitHub.client.Repositories.DownloadReleaseAsset(
		ctx, a.GitHub.GetOwner(), a.GitHub.GetRepo(), a.ReleaseAsset.GetID(), http.DefaultClient)
	if err != nil {
		return err
	}
	defer rc.Close()

	if url != "" {
		logrus.Tracef("url: %s", url)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	downloadsDir := filepath.Join(cacheDir, common.NAME, "downloads")

	filename := a.ID()

	assetFile := filepath.Join(downloadsDir, filename)
	a.DownloadPath = assetFile
	a.Extension = filepath.Ext(a.DownloadPath)

	assetFileHash := assetFile + ".sha256"

	stats, err := os.Stat(assetFileHash)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if stats != nil {
		logrus.Debugf("file already downloaded: %s", assetFile)
		return nil
	}

	// TODO: verify hash, add overwrite for force.

	hasher := sha256.New()

	// Create a temporary file
	tmpfile, err := os.Create(assetFile)
	if err != nil {
		return err
	}
	defer tmpfile.Close()

	multiWriter := io.MultiWriter(tmpfile, hasher)

	// Write the asset's content to the temporary file
	_, err = io.Copy(multiWriter, rc)
	if err != nil {
		return err
	}

	logrus.Tracef("hash: %s", fmt.Sprintf("%x", hasher.Sum(nil)))

	_ = os.WriteFile(assetFile+".sha256", []byte(fmt.Sprintf("%x", hasher.Sum(nil))), 0600)
	a.Hash = fmt.Sprintf("%s", hasher.Sum(nil))

	logrus.Tracef("Downloaded asset to: %s", tmpfile.Name())

	return nil
}
