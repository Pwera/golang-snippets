package controller

import (
	"fmt"
	"github.com/pwera/ddd/protocol/protocol"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type AuthorizationController struct {
	BaseController
	Client protocol.UserClient
}

func (ac AuthorizationController) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusBadRequest)
		return
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	user, err := ac.Client.SubmitNewUser(ctx, &protocol.NewUserRequest{
		Email: "email",
	})
	if err != nil {
		http.Error(w, "problem with SubmitNewUser", http.StatusBadRequest)
		return
	}
	fmt.Println(user)
}
