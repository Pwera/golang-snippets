package application

import "github.com/pwera/ddd/domain"

type IssueService struct {
	IssueRepository domain.IssueRepository
}

func (is IssueService) Issue(id int64) (*domain.Issue, error) { return nil, nil }
func (is IssueService) Issues() ([]*domain.Issue, error)      { return is.IssueRepository.All() }
func (is IssueService) Create(issue *domain.Issue) error {
	return is.IssueRepository.Create(issue)
}
func (is IssueService) Delete(id int64) error { return nil }
