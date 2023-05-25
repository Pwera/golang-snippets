package application

import (
	"github.com/pwera/ddd/domain"
)

type UserService struct {
	UserRepository domain.UserRepository
}

func (us UserService) Users() ([]*domain.User, error) {
	return us.UserRepository.All()
}

func (us UserService) Create(u *domain.User) error {
	return us.UserRepository.Create(u)
}
func (us UserService) Delete(id int64) error {
	return us.UserRepository.Delete(id)
}

func (us UserService) User(id int64) (*domain.User, error) {
	return us.UserRepository.User(id)
}

func (us UserService) All() ([]*domain.User, error) {
	return us.UserRepository.All()
}
