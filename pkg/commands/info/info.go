package info

import (
	"fmt"
	"github.com/apex/log"
	"github.com/ekristen/distillery/pkg/config"
	"github.com/urfave/cli/v2"
	"runtime"

	"github.com/ekristen/distillery/pkg/common"
)

func Execute(c *cli.Context) error {
	cfg, err := config.New(c.String("config"))
	if err != nil {
		return err
	}

	log.Infof("distillery/%s", common.AppVersion.Summary)
	fmt.Println("")
	log.Infof("system information")
	log.Infof("     os: %s", runtime.GOOS)
	log.Infof("   arch: %s", runtime.GOARCH)
	fmt.Println("")
	log.Infof("configuration")
	log.Infof("   home: %s", cfg.HomePath)
	log.Infof("    bin: %s", cfg.BinPath)
	log.Infof("    opt: %s", cfg.OptPath)
	log.Infof("  cache: %s", cfg.CachePath)
	fmt.Println("")
	log.Warnf("To cleanup all of distillery, remove the following directories:")
	log.Warnf("  - %s", cfg.GetCachePath())
	log.Warnf("  - %s", cfg.BinPath)
	log.Warnf("  - %s", cfg.OptPath)

	return nil
}

func Flags() []cli.Flag {
	return []cli.Flag{}
}

func init() {
	cmd := &cli.Command{
		Name:        "info",
		Usage:       "info",
		Description: `general information about distillery and the rendered configuration`,
		Flags:       append(Flags(), common.Flags()...),
		Action:      Execute,
	}

	common.RegisterCommand(cmd)
}
