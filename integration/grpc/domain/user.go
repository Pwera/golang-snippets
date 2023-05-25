package domain

type User struct {
	Id   int64
	Name string
}
type UserRepository interface {
	All() ([]*User, error)
	Create(u *User) error
	Delete(id int64) error
	User(id int64) (*User, error)
}
