package memory

import (
	"errors"

	"github.com/patrickmn/go-cache"
	"github.com/pwera/ddd/domain"
)

const (
	UsersAllKey = "Users:all"
	UserLastId  = "User:lastId"
)

var (
	ErrorEmpty    = errors.New("Empty")
	ErrorNotFound = errors.New("Not Found")
)

type UserRepository struct {
	db *cache.Cache
}

func NewUserRepository() *UserRepository {
	db := cache.New(cache.NoExpiration, cache.NoExpiration)
	db.SetDefault(UserLastId, int64(0))
	db.SetDefault(UsersAllKey, []*domain.User{})
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) All() ([]*domain.User, error) {
	result, ok := r.db.Get(UsersAllKey)
	if ok {
		return result.([]*domain.User), nil
	} else {
		return nil, ErrorEmpty
	}
}
func (r *UserRepository) getById(id int64) (*domain.User, error) {
	result, ok := r.db.Get(UsersAllKey)
	if ok {
		items := result.([]*domain.User)
		for _, user := range items {
			if user.Id == id {
				return user, nil
			}
		}
		return nil, ErrorEmpty
	}
	return nil, ErrorEmpty
}
func (r *UserRepository) Create(u *domain.User) error {
	id, _ := r.db.IncrementInt64(UserLastId, int64(1))
	u.Id = id

	result, ok := r.db.Get(UsersAllKey)
	if ok {
		result = append(result.([]*domain.User), u)
		r.db.Set(UsersAllKey, result, cache.NoExpiration)
	}
	return nil
}
func (r *UserRepository) Delete(id int64) error {
	result, ok := r.db.Get(UsersAllKey)
	if ok {
		items := result.([]*domain.User)
		for i, user := range items {
			if user.Id == id {
				items = append(items[:i], items[i+1:]...)
				return nil
			}
		}
		return ErrorNotFound
	}
	return ErrorNotFound
}

func (r *UserRepository) User(id int64) (*domain.User, error) {
	return nil, nil
}
