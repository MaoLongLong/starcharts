package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Stargazer struct {
	StarredAt time.Time `json:"starred_at"`
}

var errNoMorePages = errors.New("no more pages to get")

func (gh *GitHub) Stargazers(ctx context.Context, repo Repository) (stars []Stargazer, err error) {
	sem := make(chan bool, 4)
	var g errgroup.Group
	var lock sync.Mutex

	for page := 1; page < gh.lastPage(repo); page++ {
		sem <- true
		page := page
		g.Go(func() error {
			defer func() { <-sem }()
			result, err := gh.getStargazersPage(ctx, repo, page)
			if errors.Is(err, errNoMorePages) {
				return nil
			}
			if err != nil {
				return err
			}
			lock.Lock()
			defer lock.Unlock()
			stars = append(stars, result...)
			return nil
		})
	}

	err = g.Wait()
	sort.Slice(stars, func(i, j int) bool {
		return stars[i].StarredAt.Before(stars[j].StarredAt)
	})

	return
}

func (gh *GitHub) getStargazersPage(ctx context.Context, repo Repository, page int) ([]Stargazer, error) {
	var log = log.WithFields(log.Fields{
		"repo": repo.FullName,
		"page": page,
	})

	v, hit := gh.cache.Get(fmt.Sprintf("%s_%d", repo.FullName, page))
	if hit {
		log.Info("got from cache")
		return v.([]Stargazer), nil
	}

	var stars []Stargazer
	log.Info("getting page from api")

	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/stargazers?page=%d&per_page=%d",
		repo.FullName,
		page,
		gh.pageSize,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return stars, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3.star+json")
	if gh.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", gh.token))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return stars, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		log.Warn("rate limit hit")
		return stars, ErrRateLimit
	}

	if resp.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return stars, err
		}
		return stars, fmt.Errorf("%w: %v", ErrGitHubAPI, string(bytes))
	}

	err = json.NewDecoder(resp.Body).Decode(&stars)
	if len(stars) == 0 {
		return stars, errNoMorePages
	}

	var expire = time.Hour * 24 * 7
	if page == gh.lastPage(repo) {
		expire = time.Hour * 2
	}
	log.WithField("expire", expire).Info("caching...")

	gh.cache.Set(
		fmt.Sprintf("%s_%d", repo.FullName, page),
		stars,
		expire,
	)

	return stars, err
}

func (gh *GitHub) lastPage(repo Repository) int {
	return repo.StargazersCount/gh.pageSize + 1
}
