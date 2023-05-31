package github

import (
	"context"
	"fmt"
	"net/http"
)

const (
	reposPathWithUser = "users/%v/repos"
	defaultReposPath  = "user/repos"
)

type RepositoriesService struct {
	Client *Client
}

type Repository struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GitURL      string `json:"git_url"`
}

func (s *RepositoriesService) List(ctx context.Context, user string) ([]*Repository, *http.Response, error) {
	var path string
	if user != "" {
		path = fmt.Sprintf(reposPathWithUser, user)
	} else {
		path = defaultReposPath
	}

	req, err := s.Client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	var repos []*Repository
	resp, err := s.Client.Do(ctx, req, &repos)
	if err != nil {
		return nil, nil, err
	}
	return repos, resp, nil
}
