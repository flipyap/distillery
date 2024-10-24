package uninstall

import (
	"github.com/urfave/cli/v2"

	"github.com/ekristen/distillery/pkg/common"
)

func Execute(c *cli.Context) error {
	return nil
}

func init() {
	cmd := &cli.Command{
		Name:        "uninstall",
		Usage:       "uninstall binaries",
		Description: `uninstall binaries and all versions`,
		Before:      common.Before,
		Flags:       common.Flags(),
		Action:      Execute,
	}

	common.RegisterCommand(cmd)
}
