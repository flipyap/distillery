package osconfig

const (
	Windows = "windows"
	Linux   = "linux"
	Darwin  = "darwin"

	AMD64 = "amd64"
	ARM64 = "arm64"
)

type OS struct {
	Name          string
	Aliases       []string
	Architectures []string
	Extensions    []string
}

func (o *OS) GetOS() []string {
	return append([]string{o.Name}, o.Aliases...)
}

func (o *OS) GetArchitectures() []string {
	return o.Architectures
}

func (o *OS) GetExtensions() []string {
	return o.Extensions
}

func New(os, arch string) *OS {
	newOS := &OS{
		Name:          os,
		Architectures: []string{arch},
	}

	switch os {
	case Windows:
		newOS.Aliases = []string{"win"}
		newOS.Extensions = []string{".exe"}
	case Linux:
		newOS.Aliases = []string{}
		newOS.Extensions = []string{".AppImage"}
	case Darwin:
		newOS.Aliases = []string{"macos", "sonoma"}
		newOS.Architectures = append(newOS.Architectures, "universal")
	}

	switch arch {
	case AMD64:
		newOS.Architectures = append(newOS.Architectures, "x86_64", "64bit", "64", "x86", "64-bit", "x86-64")
	case ARM64:
		newOS.Architectures = append(newOS.Architectures, "aarch64", "armv8-a", "arm64-bit")
	}

	return newOS
}
