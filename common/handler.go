package common

import (
	"context"
	"time"
)

const (
	ModePoll  = "poll"
	ModeWatch = "watch"
)

type AgentHandler interface {
	PreHandle(ctx context.Context) error
	SetRunMode(mode string)
	PostHandle(param *HandlerParam, ctx context.Context) error
	AfterCompletion(ctx context.Context) error
}

type HandlerParam struct {
	Address  string
	Cluster  string
	ClientIp string
	AllInOne bool
	Apps     []*App
}

type App struct {
	AppId        string
	Namespaces   []string
	Secret       string
	PollInterval time.Duration
	FileName     string
	Syntax       string
}
