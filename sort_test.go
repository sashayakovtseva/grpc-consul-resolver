package consul

import (
	"sort"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
)

func TestSortSameNodeFirst(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		nodeName string
		in       []*api.ServiceEntry
		expect   []*api.ServiceEntry
	}{
		{
			name:     "one service on agent node",
			nodeName: "node-1",
			in: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "127.0.0.1",
						Port:    50051,
					},
				},
			},
			expect: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "127.0.0.1",
						Port:    50051,
					},
				},
			},
		},
		{
			name:     "one service on different node",
			nodeName: "node-1",
			in: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-2",
					},
					Service: &api.AgentService{
						Address: "127.0.0.1",
						Port:    50051,
					},
				},
			},
			expect: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-2",
					},
					Service: &api.AgentService{
						Address: "127.0.0.1",
						Port:    50051,
					},
				},
			},
		},
		{
			name:     "two services on agent node",
			nodeName: "node-1",
			in: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "227.0.0.1",
						Port:    50051,
					},
				},
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "127.0.0.1",
						Port:    50051,
					},
				},
			},
			expect: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "227.0.0.1",
						Port:    50051,
					},
				},
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "127.0.0.1",
						Port:    50051,
					},
				},
			},
		},
		{
			name:     "two services on different nodes",
			nodeName: "node-1",
			in: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "227.0.0.1",
						Port:    50051,
					},
				},
				{
					Node: &api.Node{
						Node: "node-2",
					},
					Service: &api.AgentService{
						Address: "127.0.0.1",
						Port:    50051,
					},
				},
			},
			expect: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "227.0.0.1",
						Port:    50051,
					},
				},
				{
					Node: &api.Node{
						Node: "node-2",
					},
					Service: &api.AgentService{
						Address: "127.0.0.1",
						Port:    50051,
					},
				},
			},
		},
		{
			name:     "two on same two on different",
			nodeName: "node-1",
			in: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "192.168.235.110",
						Port:    50051,
					},
				},
				{
					Node: &api.Node{
						Node: "node-2",
					},
					Service: &api.AgentService{
						Address: "192.168.235.116",
						Port:    50051,
					},
				},
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "192.168.235.112",
						Port:    50051,
					},
				},
				{
					Node: &api.Node{
						Node: "node-3",
					},
					Service: &api.AgentService{
						Address: "192.168.235.115",
						Port:    50051,
					},
				},
			},
			expect: []*api.ServiceEntry{
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "192.168.235.110",
						Port:    50051,
					},
				},
				{
					Node: &api.Node{
						Node: "node-1",
					},
					Service: &api.AgentService{
						Address: "192.168.235.112",
						Port:    50051,
					},
				},
				{
					Node: &api.Node{
						Node: "node-2",
					},
					Service: &api.AgentService{
						Address: "192.168.235.116",
						Port:    50051,
					},
				},
				{
					Node: &api.Node{
						Node: "node-3",
					},
					Service: &api.AgentService{
						Address: "192.168.235.115",
						Port:    50051,
					},
				},
			},
		},
	}

	for i := range tt {
		tc := tt[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sort.Sort(sameNodeFirst{
				agentNodeName: tc.nodeName,
				in:            tc.in,
			})

			require.Equal(t, tc.expect, tc.in)
		})
	}
}
