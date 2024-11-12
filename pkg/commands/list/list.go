package list

import (
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/distillery/pkg/common"
	"github.com/ekristen/distillery/pkg/config"
	"github.com/ekristen/distillery/pkg/inventory"
)

func Execute(c *cli.Context) error {
	cfg, err := config.New(c.String("config"))
	if err != nil {
		return err
	}

	inv := inventory.New(os.DirFS(cfg.BinPath), cfg.BinPath, cfg.GetOptPath(), cfg)

	for short, bin := range inv.Bins {
		log.Infof("%s (versions: %s)", short, strings.Join(bin.ListVersions(), ", "))
	}

	return nil
}

func init() {
	cmd := &cli.Command{
		Name:        "list",
		Usage:       "list installed binaries and versions",
		Description: `list installed binaries and versions`,
		Before:      common.Before,
		Flags:       common.Flags(),
		Action:      Execute,
	}

	common.RegisterCommand(cmd)
}
