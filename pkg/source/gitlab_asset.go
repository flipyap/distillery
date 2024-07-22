package source

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/common"
	"github.com/ekristen/distillery/pkg/source/gitlab"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type GitLabAsset struct {
	*asset.Asset

	GitLab *GitLab
	Link   *gitlab.Links
}

func (a *GitLabAsset) ID() string {
	return fmt.Sprintf("%s-%s-%s", a.GitLab.GetOwner(), a.GitLab.GetRepo(), a.GitLab.Version)
}

func (a *GitLabAsset) Download(ctx context.Context) error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	downloadsDir := filepath.Join(cacheDir, common.NAME, "downloads")
	filename := filepath.Base(a.Link.URL)

	assetFile := filepath.Join(downloadsDir, filename)
	a.DownloadPath = assetFile
	a.Extension = filepath.Ext(a.DownloadPath)

	assetFileHash := assetFile + ".sha256"
	stats, err := os.Stat(assetFileHash)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if stats != nil {
		logrus.Debug("file already downloaded")
		return nil
	}

	logrus.Infof("downloading asset: %s", a.Link.URL)

	req, err := http.NewRequest("GET", a.Link.URL, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	req.Header.Add("User-Agent", fmt.Sprintf("%s/%s", common.NAME, common.AppVersion))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	hasher := sha256.New()
	tmpFile, err := os.Create(assetFile)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	multiWriter := io.MultiWriter(tmpFile, hasher)

	f, err := os.Create(assetFile)
	if err != nil {
		return err
	}

	// Write the asset's content to the temporary file
	_, err = io.Copy(multiWriter, resp.Body)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	logrus.Tracef("hash: %s", fmt.Sprintf("%x", hasher.Sum(nil)))

	_ = os.WriteFile(assetFileHash, []byte(fmt.Sprintf("%x", hasher.Sum(nil))), 0600)
	a.Hash = fmt.Sprintf("%s", hasher.Sum(nil))

	logrus.Tracef("Downloaded asset to: %s", tmpFile.Name())

	return nil
}
