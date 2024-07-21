package list

import (
	"fmt"
	"github.com/ekristen/distillery/pkg/common"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

func Execute(c *cli.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	binDir := filepath.Join(homeDir, fmt.Sprintf(".%s", common.NAME), "bin")

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
			return nil
		}

		fmt.Println("path: ", path)
		return nil
	})

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
