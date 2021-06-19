package config

import (
	"log"

	"gopkg.in/ini.v1"
)

type config struct {
	GitHubToken    string `ini:"github-token"`
	GitHubPageSize int    `ini:"github-page-size"`
	Port           int    `ini:"port"`
}

const (
	DefaultPort           = 1113
	DefaultGitHubPageSize = 100
)

var cfg config

func init() {
	cfg.Port = DefaultPort
	cfg.GitHubPageSize = DefaultGitHubPageSize
	if err := ini.MapTo(&cfg, "starcharts.ini"); err != nil {
		log.Fatal("init config err:", err)
	}
}

func GitHubToken() string {
	return cfg.GitHubToken
}

func GitHubPageSize() int {
	return cfg.GitHubPageSize
}

func Port() int {
	return cfg.Port
}
