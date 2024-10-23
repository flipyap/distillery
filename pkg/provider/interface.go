package provider

import "context"

type ISource interface {
	GetSource() string
	GetOwner() string
	GetRepo() string
	GetApp() string
	GetID() string
	GetDownloadsDir() string
	Run(context.Context) error
}
