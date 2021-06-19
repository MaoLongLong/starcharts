package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/maolonglong/starcharts/internal/bytesconv"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

type Repository struct {
	FullName        string `json:"full_name"`
	StargazersCount int    `json:"stargazers_count"`
	CreatedAt       string `json:"created_at"`
}

func (gh *GitHub) RepoDetails(ctx context.Context, name string) (Repository, error) {
	log := log.WithField("repo", name)

	v, hit := gh.cache.Get(name)
	if hit {
		log.Info("got from cache")
		return v.(Repository), nil
	}

	var repo Repository

	url := fmt.Sprintf("https://api.github.com/repos/%s", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return repo, err
	}
	if gh.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", gh.token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return repo, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		log.Warn("rate limit hit")
		return repo, ErrRateLimit
	}

	if resp.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return repo, err
		}
		return repo, fmt.Errorf("%w: %v", ErrGitHubAPI, bytesconv.BytesToString(bytes))
	}

	err = json.NewDecoder(resp.Body).Decode(&repo)
	if err == nil {
		gh.cache.Set(name, repo, cache.DefaultExpiration)
	}

	return repo, err
}
