package source

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ekristen/distillery/pkg/asset"
)

const KubernetesSource = "kubernetes"

type Kubernetes struct {
	GitHub

	AppName string
}

func (s *Kubernetes) GetSource() string {
	return KubernetesSource
}
func (s *Kubernetes) GetOwner() string {
	return KubernetesSource
}
func (s *Kubernetes) GetRepo() string {
	return s.Repo
}
func (s *Kubernetes) GetApp() string {
	return fmt.Sprintf("%s/%s", s.Owner, s.Repo)
}
func (s *Kubernetes) GetID() string {
	return fmt.Sprintf("%s-%s", s.GetSource(), s.GetRepo())
}

func (s *Kubernetes) GetDownloadsDir() string {
	return filepath.Join(s.Options.Config.GetDownloadsPath(), s.GetSource(), s.GetOwner(), s.GetRepo(), s.Version)
}

func (s *Kubernetes) GetReleaseAssets(_ context.Context) error {
	binName := fmt.Sprintf("%s-%s-%s-%s", s.AppName, s.Version, s.GetOS(), s.GetArch())
	s.Assets = append(s.Assets, &KubernetesAsset{
		Asset:      asset.New(binName, s.AppName, s.GetOS(), s.GetArch(), s.Version),
		Kubernetes: s,
		URL: fmt.Sprintf("https://dl.k8s.io/release/v%s/bin/%s/%s/%s",
			s.Version, s.GetOS(), s.GetArch(), s.AppName),
	}, &KubernetesAsset{
		Asset:      asset.New(binName+".sha256", "", s.GetOS(), s.GetArch(), s.Version),
		Kubernetes: s,
		URL: fmt.Sprintf("https://dl.k8s.io/release/v%s/bin/%s/%s/%s.sha256",
			s.Version, s.GetOS(), s.GetArch(), s.AppName),
	})

	return nil
}

func (s *Kubernetes) Run(ctx context.Context) error {
	if err := s.sourceRun(ctx); err != nil {
		return err
	}

	// this is from the Provider struct
	if err := s.Discover([]string{s.Repo}); err != nil {
		return err
	}

	if err := s.CommonRun(ctx); err != nil {
		return err
	}

	return nil
}
