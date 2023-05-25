package main

import (
	"fmt"
	"github.com/pwera/ddd/application"
	"github.com/pwera/ddd/controller"
	"github.com/pwera/ddd/domain"
	"github.com/pwera/ddd/persistence/memory"
	"github.com/pwera/ddd/protocol/protocol"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"time"
)

type server struct{}

func (s *server) SubmitNewUser(context context.Context, ur *protocol.NewUserRequest) (*protocol.NewUserResponse, error) {
	fmt.Println("OK")
	return &protocol.NewUserResponse{Status: len(ur.Email) > 0}, nil
}
func main() {
	userRepo := memory.NewUserRepository()
	userService := application.UserService{
		UserRepository: userRepo,
	}

	userController := controller.UserController{
		UserService: userService,
	}
	for i := 0; i < 10; i += 1 {
		err := userService.Create(&domain.User{Name: fmt.Sprintf("User_%d", i)})
		if err != nil {
			fmt.Errorf("error invoking userService.Create %s", err)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", userController.List)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	s := server{}
	reflection.Register(grpcServer)
	protocol.RegisterUserServer(grpcServer, &s)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 10000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		grpcServer.Serve(lis)
	}()
	server := &http.Server{
		Addr:           ":8091",
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(server.ListenAndServe())
}
