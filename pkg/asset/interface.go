package asset

import "context"

type IAsset interface {
	GetName() string
	GetDisplayName() string
	GetScore() int
	GetType() Type
	GetAsset() *Asset
	GetFiles() []*File
	GetTempPath() string
	GetFilePath() string
	Score(options *ScoreOptions) int
	Download(context.Context) error
	Extract() error
	Install(string, string) error
	Cleanup() error
	ID() string
}
