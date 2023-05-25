package main

import (
	"fmt"
	"github.com/pwera/ddd/application"
	"github.com/pwera/ddd/controller"
	"github.com/pwera/ddd/domain"
	"github.com/pwera/ddd/persistence/db"
	"github.com/pwera/ddd/persistence/memory"
	"github.com/pwera/ddd/protocol/protocol"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

const serverAddr = "127.0.0.1:10000"

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	dial, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer dial.Close()
	userClient := protocol.NewUserClient(dial)
	_, _ = userClient.SubmitNewUser(context.Background(), nil)
	userRepo := memory.NewUserRepository()
	issueRepo := db.NewIssueRepository()
	userService := application.UserService{UserRepository: userRepo}
	issueService := application.IssueService{IssueRepository: issueRepo}
	userController := controller.UserController{UserService: userService}
	issueController := controller.IssueController{IssueService: issueService}
	authorizationController := controller.AuthorizationController{
		Client: userClient,
	}
	prepareUsers(userService)
	prepareIssues(issueService)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", userController.List)
	mux.HandleFunc("/api/issues", issueController.List)
	mux.HandleFunc("/api/register", authorizationController.Register)

	_ = http.ListenAndServe(":8090", mux)

}

func prepareIssues(issueService application.IssueService) {
	issue := domain.Issue{
		Title:       "Title",
		Priority:    domain.PriorityLow,
		OwnerId:     1,
		ProjectId:   1,
		Description: "??",
	}
	err := issueService.Create(&issue)
	if err != nil {
		fmt.Println(err)
	}
}

func prepareUsers(userService application.UserService) {
	for i := 0; i < 10; i += 1 {
		_ = userService.Create(&domain.User{Name: fmt.Sprintf("User_%d", i)})
	}
}
