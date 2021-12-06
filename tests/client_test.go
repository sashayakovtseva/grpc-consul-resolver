//go:build integration
// +build integration

package tests

import (
	"testing"
	"time"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"google.golang.org/grpc"
)

func TestClient(t *testing.T) {
	conn, err := grpc.Dial("consul://127.0.0.1:8500/whoami?wait=14s&tag=public",
		grpc.WithInsecure(),
		grpc.WithBalancerName("round_robin"),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()
	time.Sleep(29 * time.Second)
}
