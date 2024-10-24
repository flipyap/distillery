package install

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/apex/log"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/distillery/pkg/common"
	"github.com/ekristen/distillery/pkg/config"
	"github.com/ekristen/distillery/pkg/provider"
)

func Execute(c *cli.Context) error {
	cfg, err := config.New(c.String("config"))
	if err != nil {
		return err
	}

	if err := cfg.MkdirAll(); err != nil {
		return err
	}

	if c.Args().First() == "ekristen/distillery" {
		_ = c.Set("include-pre-releases", "true")
	}

	src, err := NewSource(c.Args().First(), &provider.Options{
		OS:     c.String("os"),
		Arch:   c.String("arch"),
		Config: cfg,
		Settings: map[string]interface{}{
			"version":              c.String("version"),
			"github-token":         c.String("github-token"),
			"gitlab-token":         c.String("gitlab-token"),
			"no-checksum-verify":   c.Bool("no-checksum-verify"),
			"no-score-check":       c.Bool("no-score-check"),
			"include-pre-releases": c.Bool("include-pre-releases"),
		},
	})
	if err != nil {
		return err
	}

	log.Infof("distillery/%s", common.AppVersion.Summary)
	log.Infof(" source: %s", src.GetSource())
	log.Infof("    app: %s", src.GetApp())
	log.Infof("version: %s", c.String("version"))
	log.Infof("     os: %s", c.String("os"))
	log.Infof("   arch: %s", c.String("arch"))

	if c.Bool("include-pre-releases") {
		log.Infof("including pre-releases")
	}

	if err := src.Run(c.Context); err != nil {
		return err
	}

	log.Infof("installation complete")

	return nil
}

func Before(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("no binary specified")
	}

	if c.NArg() > 1 {
		for _, arg := range c.Args().Slice() {
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("flags must be specified before the binary(ies)")
			}
		}
	}

	parts := strings.Split(c.Args().First(), "@")
	if len(parts) == 2 {
		_ = c.Set("version", parts[1])
	} else if len(parts) == 1 {
		_ = c.Set("version", "latest")
	} else {
		return fmt.Errorf("invalid binary specified")
	}

	if c.String("bin") != "" {
		_ = c.Set("bins", "false")
	}

	return common.Before(c)
}

func Flags() []cli.Flag {
	cfgDir, _ := os.UserConfigDir()

	return []cli.Flag{
		&cli.StringFlag{
			Name:  "version",
			Usage: "Specify a version to install",
			Value: "latest",
		},
		&cli.StringFlag{
			Name:     "asset",
			Usage:    "The exact name of the asset to use, useful when auto-detection fails",
			Category: "Target Selection",
		},
		&cli.StringFlag{
			Name:     "suffix",
			Usage:    "Specify the suffix to use for the binary (default is auto-detect based on OS)",
			Category: "Target Selection",
		},
		&cli.StringFlag{
			Name:     "bin",
			Usage:    "Install only the selected binary",
			Category: "Target Selection",
		},
		&cli.BoolFlag{
			Name:     "bins",
			Usage:    "Install all binaries",
			Category: "Target Selection",
			Value:    true,
		},
		&cli.StringFlag{
			Name:  "os",
			Usage: "Specify the OS to install",
			Value: runtime.GOOS,
		},
		&cli.StringFlag{
			Name:  "arch",
			Usage: "Specify the architecture to install",
			Value: runtime.GOARCH,
		},
		&cli.PathFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Specify the configuration file to use",
			EnvVars: []string{"DISTILLERY_CONFIG"},
			Value:   filepath.Join(cfgDir, fmt.Sprintf("%s.yaml", common.NAME)),
		},
		&cli.StringFlag{
			Name:     "github-token",
			Usage:    "GitHub token to use for GitHub API requests",
			EnvVars:  []string{"DISTILLERY_GITHUB_TOKEN"},
			Category: "Authentication",
		},
		&cli.StringFlag{
			Name:     "gitlab-token",
			Usage:    "GitLab token to use for GitLab API requests",
			EnvVars:  []string{"DISTILLERY_GITLAB_TOKEN"},
			Category: "Authentication",
		},
		&cli.BoolFlag{
			Name:    "include-pre-releases",
			Usage:   "Include pre-releases in the list of available versions",
			EnvVars: []string{"DISTILLERY_INCLUDE_PRE_RELEASES"},
			Aliases: []string{"pre"},
		},
		&cli.BoolFlag{
			Name:  "no-checksum-verify",
			Usage: "Disable checksum verification",
		},
		&cli.BoolFlag{
			Name:  "no-score-check",
			Usage: "Disable scoring check",
		},
	}
}

func init() {
	cmd := &cli.Command{
		Name:        "install",
		Usage:       "install [provider/]owner/repo[@version]",
		Description: fmt.Sprintf(`install a binary. default location is $HOME/.%s/bin`, common.NAME),
		Before:      Before,
		Flags:       append(Flags(), common.Flags()...),
		Action:      Execute,
		Args:        true,
		ArgsUsage:   " [provider/]owner/repo[@version]",
	}

	common.RegisterCommand(cmd)
}
