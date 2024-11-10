package inventory

import "path/filepath"

type Bin struct {
	Name     string
	Versions []*Version
	Source   string
	Owner    string
	Repo     string
}

func (b *Bin) GetInstallPath(base string) string {
	return filepath.Join(base, b.Source, b.Owner, b.Repo)
}

type Version struct {
	Version string
	Path    string
	Latest  bool
	Target  string
}
