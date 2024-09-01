package list

import (
	"fmt"
	"github.com/ekristen/distillery/pkg/common"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"strings"
)

func Execute(c *cli.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	binDir := filepath.Join(homeDir, fmt.Sprintf(".%s", common.NAME), "bin")

	bins := make(map[string]map[string]string)

	_ = filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
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

	for name, paths := range bins {
		fmt.Println("> ", name)
		for version, path := range paths {
			fmt.Println("  - ", version, path)
		}
	}

	return nil
}

func init() {
	cmd := &cli.Command{
		Name:        "list",
		Usage:       "list",
		Description: `list installed binaries`,
		Before:      common.Before,
		Flags:       common.Flags(),
		Action:      Execute,
	}

	common.RegisterCommand(cmd)
}
