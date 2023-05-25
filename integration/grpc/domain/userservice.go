package domain

type UserService interface {
	Users() ([]*User, error)
}
