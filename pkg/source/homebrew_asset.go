package source

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/common"
	"github.com/ekristen/distillery/pkg/source/homebrew"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type HomebrewAsset struct {
	*asset.Asset

	Homebrew    *Homebrew
	FileVariant *homebrew.FileVariant
}

func (a *HomebrewAsset) ID() string {
	return fmt.Sprintf("%s-%s-%s", a.Homebrew.GetOwner(), a.Homebrew.GetRepo(), a.Homebrew.Version)
}

type GHCRAuth struct {
	Token string `json:"token"`
}

func (g *GHCRAuth) Bearer() string {
	return fmt.Sprintf("Bearer %s", g.Token)
}

func (a *HomebrewAsset) getAuthToken() (*GHCRAuth, error) {
	// https://ghcr.io/token",service="ghcr.io",scope="repository:homebrew/core/ffmpeg:pull"

	req, err := http.NewRequest("GET", "https://ghcr.io/token", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("service", "ghcr.io")
	q.Add("scope", fmt.Sprintf("repository:homebrew/core/%s:%s", a.Homebrew.GetRepo(), "pull"))
	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

	var t *GHCRAuth

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}

	return t, nil
}

func (a *HomebrewAsset) Download(ctx context.Context) error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	downloadsDir := filepath.Join(cacheDir, common.NAME, "downloads")
	filename := filepath.Base(a.Name + ".tar.gz")

	assetFile := filepath.Join(downloadsDir, filename)
	a.DownloadPath = assetFile

	assetFileHash := assetFile + ".sha256"
	stats, err := os.Stat(assetFileHash)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if stats != nil {
		logrus.Debug("file already downloaded")
		return nil
	}

	token, err := a.getAuthToken()
	if err != nil {
		return err
	}

	// TODO: lookup manifest to determine how the file is stored ...

	req, err := http.NewRequest("GET", a.FileVariant.URL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", token.Bearer())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
