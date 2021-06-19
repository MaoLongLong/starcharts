package github

import (
	"errors"

	"github.com/maolonglong/starcharts/config"
	"github.com/patrickmn/go-cache"
)

var (
	ErrRateLimit = errors.New("rate limited, please try again later")
	ErrGitHubAPI = errors.New("failed to talk with github api")
)

type GitHub struct {
	token    string
	pageSize int
	cache    *cache.Cache
}

func New(cache *cache.Cache) *GitHub {
	return &GitHub{
		token:    config.GitHubToken(),
		pageSize: config.GitHubPageSize(),
		cache:    cache,
	}
}
