package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	nginxImage = "nginx"
	nginxPort  = "80/tcp"
	redisImage = "redis:latest"
	redisPort  = "6379/tcp"
)

func TestWaitForReadyToAcceptConnectionRedisLog(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        redisImage,
		ExposedPorts: []string{redisPort},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := redisC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err.Error())
		}
	}()
}

type nginxContainer struct {
	testcontainers.Container
	URI string
}

func setupNginx(ctx context.Context) (*nginxContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        nginxImage,
		ExposedPorts: []string{nginxPort},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := nginxC.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := nginxC.MappedPort(ctx, "80")
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	return &nginxContainer{Container: nginxC, URI: uri}, nil
}

func TestRunNginxContainerThatReturn200HttpStatusCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	nginxC, err := setupNginx(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := nginxC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	resp, err := http.Get(nginxC.URI)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestRunTwoParallelNginxContainers(t *testing.T) {
	ctx := context.Background()
	requests := testcontainers.ParallelContainerRequest{
		{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        nginxImage,
				ExposedPorts: []string{"10080/tcp"},
			},
			Started: true,
		}, {
			ContainerRequest: testcontainers.ContainerRequest{

				Image: nginxImage,
				ExposedPorts: []string{
					"10081/tcp",
				},
			},
			Started: true,
		},
	}

	res, err := testcontainers.ParallelContainers(ctx, requests, testcontainers.ParallelContainersOptions{})
	if err != nil {
		e, ok := err.(testcontainers.ParallelContainersError)
		if !ok {
			t.Fatalf("unknown error: %v", err)
		}

		for _, pe := range e.Errors {
			fmt.Println(pe.Request, pe.Error)
		}
		return
	}

	for _, c := range res {
		c := c
		defer func() {
			if err := c.Terminate(ctx); err != nil {
				t.Fatalf("failed to terminate container: %s", c)
			}
		}()
	}

}

func TestContainerAttachedToNewNetwork(t *testing.T) {
	aliases := []string{"alias1", "alias2", "alias3"}
	networkName := "new-network"
	ctx := context.Background()
	gcr := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        nginxImage,
			ExposedPorts: []string{nginxPort},
			Networks:     []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: aliases,
			},
		},
		Started: true,
	}

	newNetwork, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		require.NoError(t, newNetwork.Remove(ctx))
	})

	nginx, err := testcontainers.GenericContainer(ctx, gcr)

	require.NoError(t, err)
	defer func() {
		if err := nginx.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", nginx)
		}
	}()

	networks, err := nginx.Networks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networks) != 1 {
		t.Errorf("Expected networks 1. Got '%d'.", len(networks))
	}
	network := networks[0]
	if network != networkName {
		t.Errorf("Expected network name '%s'. Got '%s'.", networkName, network)
	}

	networkAliases, err := nginx.NetworkAliases(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networkAliases) != 1 {
		t.Errorf("Expected network aliases for 1 network. Got '%d'.", len(networkAliases))
	}

	networkAlias := networkAliases[networkName]

	require.NotEmpty(t, networkAlias)

	for _, alias := range aliases {
		require.Contains(t, networkAlias, alias)
	}

	networkIP, err := nginx.ContainerIP(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networkIP) == 0 {
		t.Errorf("Expected an IP address, got %v", networkIP)
	}
}

func TestRunNginxContainerWithCopiedFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	nginxC, err := setupNginx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = nginxC.CopyFileToContainer(ctx, "./testresources/hello.sh", "/hello_copy.sh", 700)
	if err != nil {
		t.Fatal(err)
	}
	_, r, err := nginxC.Exec(ctx, []string{"cat", "/hello_copy.sh"})
	if err != nil {
		t.Fatal(err)
	}
	cmdResult, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	expected := "#this is content of a hello.sh file"
	if !strings.Contains(string(cmdResult), expected) {
		t.Fatalf("File copy doesn't work, expected: %v", expected)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := nginxC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	resp, err := http.Get(nginxC.URI)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

type TestLogConsumer struct {
	Msgs []string
	Ack  chan bool
}

func (g *TestLogConsumer) Accept(l testcontainers.Log) {
	fmt.Println(string(l.Content))
	if strings.Contains(string(l.Content), "1 clients connected") {
		g.Ack <- true
		return
	}

	g.Msgs = append(g.Msgs, string(l.Content))
}

func TestFollowRedisLogsAfterSuccessfulConnection(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        redisImage,
		ExposedPorts: []string{redisPort},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
		Cmd:          []string{"--loglevel", "debug"},
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}

	ep, err := redisC.Endpoint(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	g := TestLogConsumer{
		Msgs: []string{},
		Ack:  make(chan bool),
	}

	if err = redisC.StartLogProducer(ctx); err != nil {
		t.Fatalf("Error occured when starting Log Producer with error: %v", err)
	}
	redisC.FollowOutput(&g)

	_ = redis.NewClient(&redis.Options{
		Addr: ep,
	}).Ping()

	select {
	case <-g.Ack:
	case <-time.After(5 * time.Second):
		t.Fatal("Never received final log message")
	}
	if err = redisC.StopLogProducer(); err != nil {
		t.Fatalf("Error occured when stopping Log Producer with error: %v", err)
	}

	if len(g.Msgs) != 11 {
		t.Fatalf("Log entries was not 8, instead it was:\n%v", g.Msgs)
	}

	defer func() {
		if err := redisC.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate container: %s", err.Error())
		}
	}()
}
