package provider

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/apex/log"
	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/checksum"
	"github.com/ekristen/distillery/pkg/config"
	"github.com/ekristen/distillery/pkg/cosign"
	"github.com/ekristen/distillery/pkg/osconfig"
	"github.com/ekristen/distillery/pkg/score"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

const (
	VersionLatest = "latest"
)

type Options struct {
	OS       string
	Arch     string
	Config   *config.Config
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

	ChecksumType  string
	SignatureType string
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

func (p *Provider) discoverBinary(names []string, version string) error {
	logger := logrus.WithField("discover", "binary")
	logger.Tracef("names: %v", names)

	fileScoring := map[asset.Type][]string{}
	fileScored := map[asset.Type][]score.Sorted{}

	logger.Tracef("discover: starting - %d", len(p.Assets))

	for _, a := range p.Assets {
		if _, ok := fileScoring[a.GetType()]; !ok {
			fileScoring[a.GetType()] = []string{}
		}
		fileScoring[a.GetType()] = append(fileScoring[a.GetType()], a.GetName())
	}

	for k, v := range fileScoring {
		logger.Tracef("discover: type: %d, files: %d", k, len(v))
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
			Terms:       names,
			Versions:    []string{version},
			InvalidOS:   p.OSConfig.InvalidOS(),
			InvalidArch: p.OSConfig.InvalidArchitectures(),
		})

		if len(fileScored[k]) > 0 {
			for _, vv := range fileScored[k] {
				if vv.Value >= 40 {
					highEnoughScore = true
				}
				logger.Debugf("file scoring sorted ! type: %d, scored: %v", k, vv)
			}
		}
	}

	if !highEnoughScore && !p.Options.Settings["no-score-check"].(bool) {
		logger.Error("no matching asset found, score too low")
		for _, t := range []asset.Type{asset.Binary, asset.Unknown, asset.Archive} {
			for _, v := range fileScored[t] {
				if v.Value < 40 {
					logger.Errorf("closest matching: %s (%d) (threshold: 40) -- override with --no-score-check", v.Key, v.Value)
					return errors.New("no matching asset found, score too low")
				}
			}
		}

		return errors.New("no matching asset found, score too low")
	}

	// Note: we want to look for the best binary by looking at binaries, archives and unknowns
	for _, t := range []asset.Type{asset.Binary, asset.Archive, asset.Unknown} {
		if len(fileScored[t]) > 0 {
			logger.Tracef("top scored (%d): %s (%d)", t, fileScored[t][0].Key, fileScored[t][0].Value)

			topScored := fileScored[t][0]
			if topScored.Value < 40 {
				logger.Tracef("skipped > (%d) too low: %s (%d)", t, topScored.Key, topScored.Value)
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

	return nil
}

func (p *Provider) discoverChecksum() error {
	logger := logrus.WithField("discover", "checksum")

	fileScoring := map[asset.Type][]string{}
	fileScored := map[asset.Type][]score.Sorted{}

	logger.Tracef("discover: starting - %d", len(p.Assets))

	for _, a := range p.Assets {
		if _, ok := fileScoring[a.GetType()]; !ok {
			fileScoring[a.GetType()] = []string{}
		}
		fileScoring[a.GetType()] = append(fileScoring[a.GetType()], a.GetName())
	}

	for k, v := range fileScoring {
		logger.Tracef("discover: type: %d, files: %d", k, len(v))
	}

	// Note: second pass we want to look for everything else, using binary results to help score the remaining assets
	// THis is for the checksum, signature and key files
	for k, v := range fileScoring {
		if k != asset.Checksum {
			continue
		}

		detectedOS := p.OSConfig.GetOS()
		arch := p.OSConfig.GetArchitectures()
		ext := p.OSConfig.GetExtensions()

		ext = []string{"sha256", "md5", "sha1", "txt"}
		detectedOS = []string{}
		arch = []string{}

		if _, ok := fileScored[k]; !ok {
			fileScored[k] = []score.Sorted{}
		}

		fileScored[k] = score.Score(v, &score.Options{
			OS:         detectedOS,
			Arch:       arch,
			Extensions: ext,
			WeightedTerms: map[string]int{
				"checksums": 80,
				"SHA512":    50,
				"SHA256":    40,
				"MD5":       30,
				"SHA1":      20,
				"SHA":       15,
				"SUMS":      10,
			},
			InvalidOS:   p.OSConfig.InvalidOS(),
			InvalidArch: p.OSConfig.InvalidArchitectures(),
		})

		if len(fileScored[k]) > 0 {
			for _, vv := range fileScored[k] {
				logger.Debugf("file scoring sorted ! type: %d, scored: %v", k, vv)
			}
		}
	}

	// Note: we want to look for the best binary by looking at binaries, archives and unknowns
	for _, t := range []asset.Type{asset.Checksum} {
		if len(fileScored[t]) > 0 {
			logger.Tracef("top scored (%d): %s (%d)", t, fileScored[t][0].Key, fileScored[t][0].Value)

			topScored := fileScored[t][0]
			if topScored.Value < 40 {
				logger.Tracef("skipped > (%d) too low: %s (%d)", t, topScored.Key, topScored.Value)
				continue
			}
			for _, a := range p.Assets {
				if topScored.Key == a.GetName() {
					p.Checksum = a
					break
				}
			}
		}

		if p.Checksum != nil {
			break
		}
	}

	return nil
}

func (p *Provider) determineChecksumSigTypes() error {
	logger := logrus.WithField("discover", "check-sig-type")

	p.ChecksumType = "none"
	if p.Checksum != nil {
		p.ChecksumType = p.Checksum.GetChecksumType()
	}

	p.SignatureType = "none"
	for _, a := range p.Assets {
		if a.GetType() != asset.Signature {
			continue
		}

		if p.SignatureType == "file" {
			break
		}

		if a.GetParentType() == asset.Binary || a.GetParentType() == asset.Archive || a.GetParentType() == asset.Unknown {
			p.SignatureType = "file"
		} else if a.GetParentType() == asset.Checksum {
			p.SignatureType = "checksum"
		}
	}

	logger.Tracef("checksum type: %s", p.ChecksumType)
	logger.Tracef("signature type: %s", p.SignatureType)

	return nil
}

func (p *Provider) discoverSignature(version string) error {
	logger := logrus.WithField("discover", "signature")

	fileScoring := map[asset.Type][]string{}
	fileScored := map[asset.Type][]score.Sorted{}

	logger.Tracef("discover: starting - %d", len(p.Assets))

	for _, a := range p.Assets {
		if _, ok := fileScoring[a.GetType()]; !ok {
			fileScoring[a.GetType()] = []string{}
		}
		fileScoring[a.GetType()] = append(fileScoring[a.GetType()], a.GetName())
	}

	for k, v := range fileScoring {
		logger.Tracef("discover: type: %d, files: %d", k, len(v))
	}

	var names []string
	if p.SignatureType == "checksum" {
		names = append(names, p.Checksum.GetName())
		for _, ext := range []string{"sig", "asc"} {
			names = append(names, fmt.Sprintf("%s.%s", p.Checksum.GetName(), ext))
		}
	} else if p.SignatureType == "file" {
		names = append(names, p.Binary.GetName())
		for _, ext := range []string{"sig", "asc"} {
			names = append(names, fmt.Sprintf("%s.%s", p.Binary.GetName(), ext))
		}
	}

	// Note: second pass we want to look for everything else, using binary results to help score the remaining assets
	// This is for the checksum, signature and key files
	for k, v := range fileScoring {
		if k != asset.Signature {
			continue
		}

		detectedOS := p.OSConfig.GetOS()
		arch := p.OSConfig.GetArchitectures()
		ext := p.OSConfig.GetExtensions()

		ext = []string{"sig", "asc", "sig.asc", "gpg", "keyless.sig"}
		detectedOS = []string{}
		arch = []string{}

		if _, ok := fileScored[k]; !ok {
			fileScored[k] = []score.Sorted{}
		}

		logger.Tracef("names: %v", names)

		fileScored[k] = score.Score(v, &score.Options{
			OS:          detectedOS,
			Arch:        arch,
			Extensions:  ext,
			Names:       names,
			Versions:    []string{version},
			InvalidOS:   p.OSConfig.InvalidOS(),
			InvalidArch: p.OSConfig.InvalidArchitectures(),
		})

		if len(fileScored[k]) > 0 {
			for _, vv := range fileScored[k] {
				logger.Debugf("file scoring sorted ! type: %d, scored: %v", k, vv)
			}
		}
	}

	// Note: we want to look for the best binary by looking at binaries, archives and unknowns
	for _, t := range []asset.Type{asset.Signature} {
		if len(fileScored[t]) > 0 {
			logger.Tracef("top scored (%d): %s (%d)", t, fileScored[t][0].Key, fileScored[t][0].Value)

			topScored := fileScored[t][0]
			if topScored.Value < 40 {
				logger.Tracef("skipped > (%d) too low: %s (%d)", t, topScored.Key, topScored.Value)
				continue
			}
			for _, a := range p.Assets {
				if topScored.Key == a.GetName() {
					p.Signature = a
					p.Key = a.GetMatchedAsset()
					break
				}
			}
		}

		if p.Signature != nil {
			break
		}
	}

	return nil
}

func (p *Provider) discoverKey(version string) error {
	logger := logrus.WithField("discover", "key")

	fileScoring := map[asset.Type][]string{}
	fileScored := map[asset.Type][]score.Sorted{}

	logger.Tracef("discover: starting - %d", len(p.Assets))

	for _, a := range p.Assets {
		if _, ok := fileScoring[a.GetType()]; !ok {
			fileScoring[a.GetType()] = []string{}
		}
		fileScoring[a.GetType()] = append(fileScoring[a.GetType()], a.GetName())
	}

	for k, v := range fileScoring {
		logger.Tracef("discover: type: %d, files: %d", k, len(v))
	}

	var names []string
	if p.SignatureType == "checksum" {
		names = append(names, p.Checksum.GetName())

		for _, ext := range []string{"pem", "key", "pub"} {
			names = append(names, fmt.Sprintf("%s.%s", p.Checksum.GetName(), ext))
		}
	} else if p.SignatureType == "file" {
		names = append(names, p.Binary.GetName())

		for _, ext := range []string{"pem", "key", "pub"} {
			names = append(names, fmt.Sprintf("%s.%s", p.Binary.GetName(), ext))
		}
	}

	// Note: second pass we want to look for everything else, using binary results to help score the remaining assets
	// This is for the checksum, signature and key files
	for k, v := range fileScoring {
		if k != asset.Key {
			continue
		}

		detectedOS := p.OSConfig.GetOS()
		arch := p.OSConfig.GetArchitectures()
		ext := []string{"pem", "key", "pub"}

		if _, ok := fileScored[k]; !ok {
			fileScored[k] = []score.Sorted{}
		}

		logger.Tracef("preferred names: %v", names)

		fileScored[k] = score.Score(v, &score.Options{
			OS:          detectedOS,
			Arch:        arch,
			Extensions:  ext,
			Names:       names,
			Versions:    []string{version},
			InvalidOS:   []string{},
			InvalidArch: []string{},
		})

		if len(fileScored[k]) > 0 {
			for _, vv := range fileScored[k] {
				logger.Debugf("file scoring sorted ! type: %d, scored: %v", k, vv)
			}
		}
	}

	// Note: we want to look for the best binary by looking at binaries, archives and unknowns
	for _, t := range []asset.Type{asset.Key} {
		if len(fileScored[t]) > 0 {
			logger.Tracef("top scored (%d): %s (%d)", t, fileScored[t][0].Key, fileScored[t][0].Value)

			topScored := fileScored[t][0]
			if topScored.Value < 40 {
				logger.Tracef("skipped > (%d) too low: %s (%d)", t, topScored.Key, topScored.Value)
				continue
			}
			for _, a := range p.Assets {
				if a.GetParentType() != asset.Binary &&
					a.GetParentType() != asset.Archive &&
					a.GetParentType() != asset.Unknown &&
					a.GetParentType() != asset.Checksum {
					logger.Tracef("skipping key file, not a valid parent type: %s/%s", a.GetParentType(), a.GetName())
					continue
				}

				if topScored.Key == a.GetName() {
					p.Key = a
					break
				}
			}
		}

		if p.Key != nil {
			break
		}
	}

	return nil
}

func (p *Provider) discoverMatch() error {
	logger := logrus.WithField("discover", "match")

	// Match keys to signatures.
	for _, a := range p.Assets {
		if a.GetType() != asset.Signature {
			continue
		}

		if a.GetMatchedAsset() != nil {
			continue
		}

		for _, aa := range p.Assets {
			if aa.GetType() != asset.Key {
				continue
			}

			childS := strings.TrimSuffix(aa.GetName(), filepath.Ext(aa.GetName()))
			parentS := strings.TrimSuffix(a.GetName(), filepath.Ext(a.GetName()))

			if strings.EqualFold(childS, parentS) {
				logger.Tracef("matched key: %s to signature: %s", aa.GetName(), a.GetName())
				a.SetMatchedAsset(aa)
				aa.SetMatchedAsset(a)
				break
			}
		}
	}

	// Match remaining keys to signatures, hopefully there's only a single key remaining
	// TODO: what to do if there are multiple keys remaining? (Maybe support multiple matched???)
	// Use Case: Keyless vs Keyed signing, cosign does both. The keyed file is used for multiple files.
	for _, a := range p.Assets {
		if a.GetType() != asset.Key {
			continue
		}

		if a.GetMatchedAsset() != nil {
			continue
		}

		logger.Tracef("unmatched key: %s", a.GetName())

		for _, b := range p.Assets {
			if b.GetType() != asset.Signature {
				continue
			}

			if b.GetMatchedAsset() != nil {
				continue
			}

			b.SetMatchedAsset(a)
			logger.Tracef("matched key: %s to signature: %s", a.GetName(), b.GetName())
		}

	}

	return nil
}

// Discover will attempt to discover and categorize the assets provided
func (p *Provider) Discover(names []string, version string) error { //nolint:funlen,gocyclo
	if err := p.discoverMatch(); err != nil {
		return err
	}

	if err := p.discoverBinary(names, version); err != nil {
		return err
	}

	if err := p.discoverChecksum(); err != nil {
		return err
	}

	if err := p.determineChecksumSigTypes(); err != nil {
		return err
	}

	if err := p.discoverSignature(version); err != nil {
		return err
	}

	//if err := p.discoverKey(version); err != nil {
	//	return err
	//}

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
	if p.Signature == nil {
		log.Warn("skipping signature verification (no signature)")
		return nil
	}
	if p.Key == nil {
		log.Warn("skipping signature verification (no key)")
		return nil
	}

	logrus.Trace("verifying signature")

	var fileContent []byte
	var err error
	if p.SignatureType == "checksum" {
		logrus.Trace("verifying checksum signature", p.Checksum.GetName())
		fileContent, err = os.ReadFile(p.Checksum.GetFilePath())
		if err != nil {
			return err
		}
	} else {
		logrus.Trace("verifying binary signature")
		fileContent, err = os.ReadFile(p.Binary.GetFilePath())
		if err != nil {
			return err
		}
	}

	logrus.Trace("key file name: ", p.Key.GetName())

	publicKeyContentEncoded, err := os.ReadFile(p.Key.GetFilePath())
	if err != nil {
		return err
	}

	publicKeyContent, err := base64.StdEncoding.DecodeString(string(publicKeyContentEncoded))
	if err != nil {
		if errors.Is(err, base64.CorruptInputError(0)) {
			publicKeyContent = publicKeyContentEncoded
		} else {
			return err
		}
	}

	pubKey, err := cosign.ParsePublicKey(publicKeyContent)
	if err != nil {
		return err
	}

	logrus.Trace("signature file name: ", p.Signature.GetName())

	sigData, err := os.ReadFile(p.Signature.GetFilePath())
	if err != nil {
		return err
	}

	dataHash := cosign.HashData(fileContent)

	valid, err := cosign.VerifySignature(pubKey, dataHash, sigData)
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("unable to validate signature")
	}

	log.Info("signature verified")

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
	return p.Binary.Install(
		p.Binary.ID(), p.Options.Config.BinPath, filepath.Join(p.Options.Config.OptPath, p.Binary.Path()))
}

func (p *Provider) Cleanup() error {
	return p.Binary.Cleanup()
}
