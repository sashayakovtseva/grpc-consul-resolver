package consul

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
)

func TestResolver_WatchConsulService(t *testing.T) {
	tt := []struct {
		name   string
		target *target
		setup  func(m *MockConsul)
		expect []*api.ServiceEntry
	}{
		{
			name: "ok no limit",
			target: &target{
				Service: "svc",
				Wait:    time.Second,
				Healthy: true,
				Near:    "_agent",
			},
			setup: func(m *MockConsul) {
				m.EXPECT().ServiceMultipleTags("svc", nil, true, &api.QueryOptions{
					Near:     "_agent",
					WaitTime: time.Second,
				}).Return([]*api.ServiceEntry{
					{
						Service: &api.AgentService{Address: "127.0.0.1", Port: 1024},
					},
				}, &api.QueryMeta{LastIndex: 1}, nil)

				m.EXPECT().ServiceMultipleTags("svc", nil, true, &api.QueryOptions{
					WaitIndex: 1,
					Near:      "_agent",
					WaitTime:  time.Second,
				}).DoAndReturn(func(
					_ string,
					_ []string,
					_ bool,
					opt *api.QueryOptions,
				) ([]*api.ServiceEntry, *api.QueryMeta, error) {
					select {}
				})
			},
			expect: []*api.ServiceEntry{
				{
					Service: &api.AgentService{Address: "127.0.0.1", Port: 1024},
				},
			},
		},
		{
			name: "ok limit sort same node first",
			target: &target{
				Service: "svc",
				Wait:    time.Second,
				Near:    "_agent",
				tags:    []string{"green", "master"},
				Limit:   1,
				Sort:    sortSameNodeFirst,
			},
			setup: func(m *MockConsul) {
				m.EXPECT().ServiceMultipleTags("svc", []string{"green", "master"}, false, &api.QueryOptions{
					WaitTime: time.Second,
					Near:     "_agent",
				}).Return([]*api.ServiceEntry{
					{Node: &api.Node{Node: "myNode2"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 1024}},
					{Node: &api.Node{Node: "myNode"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 8080}},
					{Node: &api.Node{Node: "myNode2"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 1025}},
					{Node: &api.Node{Node: "myNode"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 8081}},
					{Node: &api.Node{Node: "myNode2"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 1026}},
				}, &api.QueryMeta{LastIndex: 1}, nil)

				m.EXPECT().ServiceMultipleTags("svc", []string{"green", "master"}, false, &api.QueryOptions{
					WaitIndex: 1,
					Near:      "_agent",
					WaitTime:  time.Second,
				}).DoAndReturn(func(
					_ string,
					_ []string,
					_ bool,
					opt *api.QueryOptions,
				) ([]*api.ServiceEntry, *api.QueryMeta, error) {
					select {}
				})
			},
			expect: []*api.ServiceEntry{
				{Node: &api.Node{Node: "myNode"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 8080}},
			},
		},
		{
			name: "consul error",
			target: &target{
				Service: "svc",
				Wait:    time.Second,
				Near:    "_agent",
				tags:    []string{"green", "master"},
				Limit:   1,
				Sort:    sortByName,
			},
			setup: func(m *MockConsul) {
				m.EXPECT().ServiceMultipleTags("svc", []string{"green", "master"}, false, &api.QueryOptions{
					WaitTime: time.Second,
					Near:     "_agent",
				}).Return(nil, nil, fmt.Errorf("some error"))

				m.EXPECT().ServiceMultipleTags("svc", []string{"green", "master"}, false, &api.QueryOptions{
					WaitTime: time.Second,
					Near:     "_agent",
				}).Return([]*api.ServiceEntry{
					{Node: &api.Node{Node: "myNode2"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 1024}},
					{Node: &api.Node{Node: "myNode"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 8080}},
					{Node: &api.Node{Node: "myNode2"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 1025}},
					{Node: &api.Node{Node: "myNode"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 8081}},
					{Node: &api.Node{Node: "myNode2"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 1026}},
				}, &api.QueryMeta{LastIndex: 1}, nil)

				m.EXPECT().ServiceMultipleTags("svc", []string{"green", "master"}, false, &api.QueryOptions{
					WaitIndex: 1,
					Near:      "_agent",
					WaitTime:  time.Second,
				}).DoAndReturn(func(
					_ string,
					_ []string,
					_ bool,
					opt *api.QueryOptions,
				) ([]*api.ServiceEntry, *api.QueryMeta, error) {
					select {}
				})
			},
			expect: []*api.ServiceEntry{
				{Node: &api.Node{Node: "myNode2"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 1024}},
			},
		},
		{
			name: "no change",
			target: &target{
				Service: "svc",
				Wait:    time.Second,
				Near:    "_agent",
				tags:    []string{"green", "master"},
				Limit:   1,
				Sort:    sortByName,
			},
			setup: func(m *MockConsul) {
				m.EXPECT().ServiceMultipleTags("svc", []string{"green", "master"}, false, &api.QueryOptions{
					WaitTime: time.Second,
					Near:     "_agent",
				}).Return(nil, &api.QueryMeta{}, nil)

				m.EXPECT().ServiceMultipleTags("svc", []string{"green", "master"}, false, &api.QueryOptions{
					WaitTime: time.Second,
					Near:     "_agent",
				}).Return([]*api.ServiceEntry{
					{Node: &api.Node{Node: "myNode2"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 1024}},
				}, &api.QueryMeta{LastIndex: 1}, nil)

				m.EXPECT().ServiceMultipleTags("svc", []string{"green", "master"}, false, &api.QueryOptions{
					WaitIndex: 1,
					Near:      "_agent",
					WaitTime:  time.Second,
				}).DoAndReturn(func(
					_ string,
					_ []string,
					_ bool,
					opt *api.QueryOptions,
				) ([]*api.ServiceEntry, *api.QueryMeta, error) {
					select {}
				})
			},
			expect: []*api.ServiceEntry{
				{Node: &api.Node{Node: "myNode2"}, Service: &api.AgentService{Address: "127.0.0.1", Port: 1024}},
			},
		},
	}
	for i := range tt {
		tc := tt[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockConsul := NewMockConsul(ctrl)
			if tc.setup != nil {
				tc.setup(mockConsul)
			}

			s := &Resolver{
				logger:        noopLogger{},
				t:             tc.target,
				c:             mockConsul,
				agentNodeName: "myNode",
			}

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			out := s.WatchServiceChanges(ctx)
			time.Sleep(5 * time.Millisecond)

			select {
			case <-ctx.Done():
				return
			case got := <-out:
				require.Equal(t, tc.expect, got)
			}
		})
	}
}
