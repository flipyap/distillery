package provider

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/checksum"
	"github.com/ekristen/distillery/pkg/cosign"
	"github.com/ekristen/distillery/pkg/osconfig"
	"github.com/ekristen/distillery/pkg/score"
)

const (
	VersionLatest = "latest"
)

type Options struct {
	OS           string
	Arch         string
	HomeDir      string
	CacheDir     string
	BinDir       string
	OptDir       string
	MetadataDir  string
	DownloadsDir string

	Settings map[string]interface{}
}

type Provider struct {
	Options   *Options
	OSConfig  *osconfig.OS
	Assets    []asset.IAsset
	Binary    asset.IAsset
	Signature asset.IAsset
	Checksum  asset.IAsset
	Key       asset.IAsset
}

func (p *Provider) GetOS() string {
	return p.Options.OS
}

func (p *Provider) GetArch() string {
	return p.Options.Arch
}

// CommonRun - common run logic for all sources that includes download, extract, install and cleanup
func (p *Provider) CommonRun(ctx context.Context) error {
	if err := p.Download(ctx); err != nil {
		return err
	}

	defer func(s *Provider) {
		err := s.Cleanup()
		if err != nil {
			log.WithError(err).Error("unable to cleanup")
		}
	}(p)

	if err := p.Verify(); err != nil {
		return err
	}

	if err := p.Extract(); err != nil {
		return err
	}

	if err := p.Install(); err != nil {
		return err
	}

	return nil
}

// Discover will attempt to discover and categorize the assets provided
// TODO(ek): split up and refactor this function as it's way too complex
func (p *Provider) Discover(names []string) error { //nolint:funlen,gocyclo
	fileScoring := map[asset.Type][]string{}
	fileScored := map[asset.Type][]score.Sorted{}

	logrus.Tracef("discover: starting - %d", len(p.Assets))

	for _, a := range p.Assets {
		if _, ok := fileScoring[a.GetType()]; !ok {
			fileScoring[a.GetType()] = []string{}
		}
		fileScoring[a.GetType()] = append(fileScoring[a.GetType()], a.GetName())
	}

	for k, v := range fileScoring {
		logrus.Tracef("discover: type: %d, files: %d", k, len(v))
	}

	highEnoughScore := false

	// Note: first pass we want to look for just binaries, archives and unknowns and score and sort them
	for k, v := range fileScoring {
		if k != asset.Binary && k != asset.Unknown && k != asset.Archive {
			continue
		}

		detectedOS := p.OSConfig.GetOS()
		arch := p.OSConfig.GetArchitectures()
		ext := p.OSConfig.GetExtensions()

		if _, ok := fileScored[k]; !ok {
			fileScored[k] = []score.Sorted{}
		}

		fileScored[k] = score.Score(v, &score.Options{
			OS:          detectedOS,
			Arch:        arch,
			Extensions:  ext,
			Names:       names,
			InvalidOS:   p.OSConfig.InvalidOS(),
			InvalidArch: p.OSConfig.InvalidArchitectures(),
		})

		if len(fileScored[k]) > 0 {
			for _, vv := range fileScored[k] {
				if vv.Value >= 40 {
					highEnoughScore = true
				}
				logrus.Debugf("file scoring sorted ! type: %d, scored: %v", k, vv)
			}
		}
	}

	if !highEnoughScore && !p.Options.Settings["no-score-check"].(bool) {
		log.Error("no matching asset found, score too low")
		for _, t := range []asset.Type{asset.Binary, asset.Unknown, asset.Archive} {
			for _, v := range fileScored[t] {
				if v.Value < 40 {
					log.Errorf("closest matching: %p (%d) (threshold: 40) -- override with --no-score-check", v.Key, v.Value)
					return errors.New("no matching asset found, score too low")
				}
			}
		}

		return errors.New("no matching asset found, score too low")
	}

	// Note: we want to look for the best binary by looking at binaries, archives and unknowns
	for _, t := range []asset.Type{asset.Binary, asset.Archive, asset.Unknown} {
		if len(fileScored[t]) > 0 {
			logrus.Tracef("top scored (%d): %s (%d)", t, fileScored[t][0].Key, fileScored[t][0].Value)

			topScored := fileScored[t][0]
			if topScored.Value < 40 {
				logrus.Tracef("skipped > (%d) too low: %s (%d)", t, topScored.Key, topScored.Value)
				continue
			}
			for _, a := range p.Assets {
				if topScored.Key == a.GetName() {
					p.Binary = a
					break
				}
			}
		}

		if p.Binary != nil {
			break
		}
	}

	if p.Binary == nil {
		return errors.New("no binary found")
	}

	// Note: second pass we want to look for everything else, using binary results to help score the remaining assets
	// THis is for the checksum, signature and key files
	for k, v := range fileScoring {
		if k == asset.Binary || k == asset.Unknown || k == asset.Archive {
			continue
		}

		detectedOS := p.OSConfig.GetOS()
		arch := p.OSConfig.GetArchitectures()
		ext := p.OSConfig.GetExtensions()

		if k == asset.Key {
			ext = []string{"key", "pub", "pem"}
			detectedOS = []string{}
			arch = []string{}
		} else if k == asset.Signature {
			ext = []string{"sig", "asc"}
			detectedOS = []string{}
			arch = []string{}
		} else if k == asset.Checksum {
			ext = []string{"sha256", "md5", "sha1", "txt"}
			detectedOS = []string{}
			arch = []string{}
		}

		if _, ok := fileScored[k]; !ok {
			fileScored[k] = []score.Sorted{}
		}

		fileScored[k] = score.Score(v, &score.Options{
			OS:          detectedOS,
			Arch:        arch,
			Extensions:  ext,
			Names:       []string{strings.ReplaceAll(p.Binary.GetName(), filepath.Ext(p.Binary.GetName()), "")},
			InvalidOS:   p.OSConfig.InvalidOS(),
			InvalidArch: p.OSConfig.InvalidArchitectures(),
		})

		if len(fileScored[k]) > 0 {
			logrus.Debugf("file scoring sorted ! type: %d, scored: %v", k, fileScored[k][0])
		}
	}

	for _, a := range p.Assets {
		for k, v := range fileScored {
			vv := v[0]

			if a.GetType() == asset.Checksum && a.GetType() == k && a.GetName() == vv.Key { //nolint:gocritic
				p.Checksum = a
			}
			if a.GetType() == asset.Signature && a.GetType() == k && a.GetName() == vv.Key { //nolint:gocritic
				p.Signature = a
			}
			if a.GetType() == asset.Key && a.GetType() == k && a.GetName() == vv.Key { //nolint:gocritic
				p.Key = a
			}
		}
	}

	if p.Binary != nil {
		logrus.Tracef("best binary: %s", p.Binary.GetName())
	}
	if p.Checksum != nil {
		logrus.Tracef("best checksum: %s", p.Checksum.GetName())
	}
	if p.Signature != nil {
		logrus.Tracef("best signature: %s", p.Signature.GetName())
	}
	if p.Key != nil {
		logrus.Tracef("best key: %s", p.Key.GetName())
	}

	return nil
}

func (p *Provider) Download(ctx context.Context) error {
	log.Info("downloading assets")
	if p.Binary != nil {
		if err := p.Binary.Download(ctx); err != nil {
			return err
		}
	}

	if p.Signature != nil {
		if err := p.Signature.Download(ctx); err != nil {
			return err
		}
	}

	if p.Checksum != nil {
		if err := p.Checksum.Download(ctx); err != nil {
			return err
		}
	}

	if p.Key != nil {
		if err := p.Key.Download(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) Verify() error {
	if err := p.verifyChecksum(); err != nil {
		return err
	}

	return p.verifySignature()
}

func (p *Provider) verifySignature() error {
	if true {
		log.Debug("skipping signature verification")
		return nil
	}

	logrus.Info("verifying signature")

	cosignFileContent, err := os.ReadFile(p.Checksum.GetFilePath())
	if err != nil {
		return err
	}

	publicKeyContentEncoded, err := os.ReadFile(p.Key.GetFilePath())
	if err != nil {
		return err
	}

	publicKeyContent, err := base64.StdEncoding.DecodeString(string(publicKeyContentEncoded))
	if err != nil {
		return err
	}

	pubKey, err := cosign.ParsePublicKey(publicKeyContent)
	if err != nil {
		return err
	}

	fmt.Printf("Public Key: %+v\n", pubKey)

	sigData, err := os.ReadFile(p.Signature.GetFilePath())
	if err != nil {
		return err
	}

	valid, err := cosign.VerifySignature(pubKey, cosignFileContent, sigData)
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("unable to validate signature")
	}

	return nil
}

// verifyChecksum - verify the checksum of the binary
func (p *Provider) verifyChecksum() error {
	if v, ok := p.Options.Settings["no-checksum-verify"]; ok && v.(bool) {
		log.Warn("skipping checksum verification")
		return nil
	}

	if p.Checksum == nil {
		log.Warn("skipping checksum verification (no checksum)")
		return nil
	}

	logrus.Debug("verifying checksum")
	logrus.Tracef("binary: %s", p.Binary.GetName())

	match, err := checksum.CompareHashWithChecksumFile(p.Binary.GetName(),
		p.Binary.GetFilePath(), p.Checksum.GetFilePath(), sha256.New)
	if err != nil {
		return err
	}

	logrus.Tracef("checksum match: %v", match)

	if !match {
		return fmt.Errorf("checksum verification failed")
	}

	log.Info("checksum verified")

	return nil
}

func (p *Provider) Extract() error {
	return p.Binary.Extract()
}

func (p *Provider) Install() error {
	return p.Binary.Install(p.Binary.ID(), p.Options.BinDir)
}

func (p *Provider) Cleanup() error {
	return p.Binary.Cleanup()
}
