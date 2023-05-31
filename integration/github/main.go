package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

	"github.com/pwera/github/github"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("GITHUB_TOKEN")
	user := os.Getenv("REPO_USER")
	service := github.RepositoriesService{Client: github.NewClient(context.Background(), token)}
	list, _, err := service.List(context.Background(), user)
	if err != nil {
		log.Print(err)
	}
	for i, r := range list {
		fmt.Printf("%d: %v\n", i, r)
	}
}
