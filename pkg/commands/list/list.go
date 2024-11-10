package list

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/distillery/pkg/common"
	"github.com/ekristen/distillery/pkg/config"
)

func Execute(c *cli.Context) error {
	cfg, err := config.New(c.String("config"))
	if err != nil {
		return err
	}

	bins := make(map[string]map[string]string)

	_ = filepath.Walk(cfg.BinPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileInfo, err := os.Lstat(path)
		if err != nil {
			return err
		}

		if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			simpleName := info.Name()
			version := "latest"
			parts := strings.Split(info.Name(), "@")
			if len(parts) > 1 {
				simpleName = parts[0]
				version = parts[1]
			}

			if _, ok := bins[simpleName]; !ok {
				bins[simpleName] = make(map[string]string)
			}

			bins[simpleName][version] = path

			return nil
		}

		return nil
	})

	var keys []string
	for key := range bins {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		var versions []string
		for version := range bins[key] {
			versions = append(versions, version)
		}
		log.Infof("%s (versions: %s)", key, strings.Join(versions, ", "))
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
