package consul

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

func TestPopulateEndpoints(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name  string
		input []*api.ServiceEntry
		want  []resolver.Address
	}{
		{
			name: "one",
			input: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address:     "127.0.0.1",
						Port:        50051,
						CreateIndex: 100,
						ModifyIndex: 100,
					},
				},
			},
			want: []resolver.Address{
				{
					Addr: "127.0.0.1:50051",
				},
			},
		},
		{
			name: "two",
			input: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address:     "127.0.0.1",
						Port:        50051,
						CreateIndex: 100,
						ModifyIndex: 100,
					},
				},
				{
					Node: &api.Node{
						Node: "node-2",
					},
					Service: &api.AgentService{
						Address:     "227.0.0.1",
						Port:        50051,
						CreateIndex: 101,
						ModifyIndex: 101,
					},
				},
			},
			want: []resolver.Address{
				{
					Addr: "127.0.0.1:50051",
				},
				{
					Addr: "227.0.0.1:50051",
				},
			},
		},
	}
	for i := range tt {
		tc := tt[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			clientConnMock := NewMockClientConn(ctrl)
			clientConnMock.EXPECT().UpdateState(resolver.State{Addresses: tc.want})

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			in := make(chan []*api.ServiceEntry, 1)
			in <- tc.input

			go populateEndpoints(ctx, clientConnMock, in)

			time.Sleep(time.Millisecond)
		})
	}
}
