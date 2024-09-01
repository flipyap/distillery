package score

import (
	"testing"

	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	logrus.SetLevel(logrus.TraceLevel)
}

func TestScore(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		names    []string
		opts     *Options
		expected []Sorted
	}{
		{
			name:  "unsupported extension",
			names: []string{"dist-linux-amd64.deb"},
			opts: &Options{
				OS:         []string{"linux"},
				Arch:       []string{"amd64"},
				Extensions: []string{"unknown"},
			},
			expected: []Sorted{
				{
					Key:   "dist-linux-amd64.deb",
					Value: 15,
				},
			},
		},
		{
			name: "simple binary",
			names: []string{
				"dist-linux-amd64",
			},
			opts: &Options{
				OS:   []string{"linux"},
				Arch: []string{"amd64"},
				Extensions: []string{
					matchers.TypeGz.Extension,
					types.Unknown.Extension,
					matchers.TypeZip.Extension,
					matchers.TypeXz.Extension,
					matchers.TypeTar.Extension,
					matchers.TypeBz2.Extension,
					matchers.TypeExe.Extension,
				},
			},
			expected: []Sorted{
				{
					Key:   "dist-linux-amd64",
					Value: 15,
				},
			},
		},
		{
			name: "simple binary matching signature file",
			names: []string{
				"dist-linux-amd64.sig",
			},
			opts: &Options{
				OS:         []string{"linux"},
				Arch:       []string{"amd64"},
				Extensions: []string{"sig"},
				Names:      []string{"dist"},
			},
			expected: []Sorted{
				{
					Key:   "dist-linux-amd64.sig",
					Value: 50,
				},
			},
		},
		{
			name: "simple binary matching key file",
			names: []string{
				"dist-linux-amd64.pem",
			},
			opts: &Options{
				OS:         []string{"linux"},
				Arch:       []string{"amd64"},
				Extensions: []string{"pem", "pub"},
			},
			expected: []Sorted{
				{
					Key:   "dist-linux-amd64.pem",
					Value: 45,
				},
			},
		},
		{
			name: "global checksums file",
			names: []string{
				"checksums.txt",
				"SHA256SUMS",
				"SHASUMS",
			},
			opts: &Options{
				OS:         []string{"linux"},
				Arch:       []string{"amd64"},
				Extensions: []string{"txt"},
			},
			expected: []Sorted{
				{
					Key:   "checksums.txt",
					Value: 15,
				},
				{
					Key:   "SHA256SUMS",
					Value: 0,
				},
				{
					Key:   "SHASUMS",
					Value: 0,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := Score(c.names, c.opts)
			assert.Equal(t, c.expected, actual)
		})
	}
}
