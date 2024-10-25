package install

import (
	"fmt"
	"strings"

	"github.com/ekristen/distillery/pkg/osconfig"
	"github.com/ekristen/distillery/pkg/provider"
	"github.com/ekristen/distillery/pkg/source"
)

func NewSource(src string, opts *provider.Options) (provider.ISource, error) {
	detectedOS := osconfig.New(opts.OS, opts.Arch)

	version := "latest"
	versionParts := strings.Split(src, "@")
	if len(versionParts) > 1 {
		src = versionParts[0]
		version = versionParts[1]
	}

	parts := strings.Split(src, "/")

	if len(parts) == 1 {
		switch opts.Config.DefaultSource {
		case source.HomebrewSource:
			return &source.Homebrew{
				Provider: provider.Provider{Options: opts, OSConfig: detectedOS},
				Formula:  parts[0],
				Version:  version,
			}, nil
		case source.HashicorpSource:
			return &source.Hashicorp{
				Provider: provider.Provider{Options: opts, OSConfig: detectedOS},
				Owner:    parts[0],
				Repo:     parts[0],
				Version:  version,
			}, nil
		}

		return nil, fmt.Errorf("invalid install source, expect format of owner/repo or owner/repo@version")
	}

	if len(parts) == 2 {
		// could be GitHub or Homebrew or Hashicorp
		if parts[0] == source.HomebrewSource {
			return &source.Homebrew{
				Provider: provider.Provider{Options: opts, OSConfig: detectedOS},
				Formula:  parts[1],
				Version:  version,
			}, nil
		} else if parts[0] == source.HashicorpSource {
			return &source.Hashicorp{
				Provider: provider.Provider{Options: opts, OSConfig: detectedOS},
				Owner:    parts[1],
				Repo:     parts[1],
				Version:  version,
			}, nil
		}

		switch opts.Config.DefaultSource {
		case source.GitHubSource:
			return &source.GitHub{
				Provider: provider.Provider{Options: opts, OSConfig: detectedOS},
				Owner:    parts[0],
				Repo:     parts[1],
				Version:  version,
			}, nil
		case "gitlab":
			return &source.GitLab{
				Provider: provider.Provider{Options: opts, OSConfig: detectedOS},
				Owner:    parts[0],
				Repo:     parts[1],
				Version:  version,
			}, nil
		}

		return nil, fmt.Errorf("invalid install source, expect format of owner/repo or owner/repo@version")
	} else if len(parts) >= 3 {
		if strings.HasPrefix(parts[0], "github") {
			if parts[1] == source.HashicorpSource {
				return &source.Hashicorp{
					Provider: provider.Provider{Options: opts, OSConfig: detectedOS},
					Owner:    parts[1],
					Repo:     parts[2],
					Version:  version,
				}, nil
			}

			return &source.GitHub{
				Provider: provider.Provider{Options: opts, OSConfig: detectedOS},
				Owner:    parts[1],
				Repo:     parts[2],
				Version:  version,
			}, nil
		} else if strings.HasPrefix(parts[0], "gitlab") {
			return &source.GitLab{
				Provider: provider.Provider{Options: opts, OSConfig: detectedOS},
				Owner:    parts[1],
				Repo:     parts[2],
				Version:  version,
			}, nil
		}

		return nil, fmt.Errorf("unknown source: %s", src)
	}

	return nil, fmt.Errorf("unknown source: %s", src)
}
