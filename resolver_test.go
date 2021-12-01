package consul

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
)

func TestResolver_WatchConsulService(t *testing.T) {
	tests := []struct {
		name             string
		tgt              *target
		services         []*api.ServiceEntry
		errorFromService error
		want             []*api.ServiceEntry
	}{
		{
			name: "no limit",
			tgt: &target{
				Service: "svc",
				Wait:    time.Second,
			},
			services: []*api.ServiceEntry{
				{
					Service: &api.AgentService{Address: "127.0.0.1", Port: 1024},
				},
			},
			want: []*api.ServiceEntry{
				{
					Service: &api.AgentService{Address: "127.0.0.1", Port: 1024},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			ctrl := gomock.NewController(t)

			fconsul := NewMockConsul(ctrl)
			fconsul.EXPECT().ServiceMultipleTags(tt.tgt.Service, tt.tgt.tags, tt.tgt.Healthy, &api.QueryOptions{
				WaitIndex:         0,
				Near:              tt.tgt.Near,
				WaitTime:          tt.tgt.Wait,
				Datacenter:        tt.tgt.Dc,
				AllowStale:        tt.tgt.AllowStale,
				RequireConsistent: tt.tgt.RequireConsistent,
			}).Return(tt.services, &api.QueryMeta{LastIndex: 1}, tt.errorFromService).Times(1)

			fconsul.EXPECT().ServiceMultipleTags(tt.tgt.Service, tt.tgt.tags, tt.tgt.Healthy, &api.QueryOptions{
				WaitIndex:         1,
				Near:              tt.tgt.Near,
				WaitTime:          tt.tgt.Wait,
				Datacenter:        tt.tgt.Dc,
				AllowStale:        tt.tgt.AllowStale,
				RequireConsistent: tt.tgt.RequireConsistent,
			}).DoAndReturn(func(_ string, _ []string, _ bool, opt *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error) {
				if opt.WaitIndex > 0 {
					select {}
				}
				return tt.services, &api.QueryMeta{LastIndex: 1}, tt.errorFromService
			}).Times(1)

			s := &Resolver{
				logger: noopLogger{},
				t:      tt.tgt,
				c:      fconsul,
			}

			out := s.WatchServiceChanges(ctx)
			time.Sleep(5 * time.Millisecond)

			select {
			case <-ctx.Done():
				return
			case got := <-out:
				require.Equal(t, tt.want, got)
			}
		})
	}
}
