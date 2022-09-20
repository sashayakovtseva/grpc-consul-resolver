//go:build integration

package tests

import (
	"testing"
	"time"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestClient(t *testing.T) {
	conn, err := grpc.Dial("consul://127.0.0.1:8500/whoami?wait=14s&tag=public",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()
	time.Sleep(29 * time.Second)
}
