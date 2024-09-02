package source_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ekristen/distillery/pkg/source"
)

func Test_New(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		want   source.ISource
	}{
		{
			source: "ekristen/aws-nuke",
			want: &source.GitHub{
				Owner:   "ekristen",
				Repo:    "aws-nuke",
				Version: "latest",
			},
		},
		{
			source: "github/ekristen/aws-nuke",
			want: &source.GitHub{
				Owner:   "ekristen",
				Repo:    "aws-nuke",
				Version: "latest",
			},
		},
		{
			source: "github.com/ekristen/aws-nuke",
			want: &source.GitHub{
				Owner:   "ekristen",
				Repo:    "aws-nuke",
				Version: "latest",
			},
		},
		{
			source: "ekristen/aws-nuke@3.1.1",
			want: &source.GitHub{
				Owner:   "ekristen",
				Repo:    "aws-nuke",
				Version: "3.1.1",
			},
		},
		{
			source: "github/ekristen/aws-nuke@3.1.1",
			want: &source.GitHub{
				Owner:   "ekristen",
				Repo:    "aws-nuke",
				Version: "3.1.1",
			},
		},
		{
			source: "github.com/ekristen/aws-nuke@3.1.1",
			want: &source.GitHub{
				Owner:   "ekristen",
				Repo:    "aws-nuke",
				Version: "3.1.1",
			},
		},
		{
			source: "homebrew/aws-nuke",
			want: &source.Homebrew{
				Formula: "aws-nuke",
				Version: "latest",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.source, func(t *testing.T) {
			got, err := source.New(tt.source, &source.Options{})
			assert.NoError(t, err)
			assert.Equal(t, tt.want.GetSource(), got.GetSource())
		})
	}
}
