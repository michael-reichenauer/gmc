package client

import (
	"github.com/michael-reichenauer/gmc/api"
)

type client struct {
	server api.Api
}

func NewClient(server api.Api) api.Api {
	return &client{server: server}
}

func (t client) OpenRepo(path string) (api.Repo, error) {
	return t.server.OpenRepo(path)
}

func (t client) GetRecentDirs() ([]string, error) {
	return t.server.GetRecentDirs()
}

func (t client) GetSubDirs(path string) ([]string, error) {
	return t.server.GetSubDirs(path)
}
