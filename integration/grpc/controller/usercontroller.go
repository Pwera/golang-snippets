package controller

import (
	"net/http"

	"github.com/pwera/ddd/domain"
)

type UserController struct {
	BaseController
	UserService domain.UserService
}

func (c UserController) List(w http.ResponseWriter, r *http.Request) {
	users, err := c.UserService.Users()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.MarshalAndWriteHeaders(users, w)
}

func (c UserController) Show(w http.ResponseWriter, r *http.Request)   {}
func (c UserController) Create(w http.ResponseWriter, r *http.Request) {}
func (c UserController) Delete(w http.ResponseWriter, r *http.Request) {}
