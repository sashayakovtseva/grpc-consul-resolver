package consul

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/jpillora/backoff"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

// init function for resolver registration.
func init() {
	resolver.Register(&builder{})
}

// resolvr implements resolver.Resolver from the gRPC package.
// It watches for endpoints changes and pushes them to the underlying gRPC connection.
type resolvr struct {
	cancelFunc context.CancelFunc
}

// ResolveNow will be skipped due unnecessary in this case
func (r *resolvr) ResolveNow(resolver.ResolveNowOptions) {}

// Close closes the resolver.
func (r *resolvr) Close() {
	r.cancelFunc()
}

//go:generate mockgen -package mocks -destination internal/mocks/resolverClientConn.go  google.golang.org/grpc/resolver ClientConn
//go:generate mockgen -package mocks -destination internal/mocks/servicer.go -source consul.go servicer
type servicer interface {
	ServiceMultipleTags(service string, tags []string, passingOnly bool, q *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error)
}

func watchConsulService(ctx context.Context, s servicer, tgt target, out chan<- []*api.ServiceEntry) {
	bck := &backoff.Backoff{
		Factor: 2,
		Jitter: true,
		Min:    10 * time.Millisecond,
		Max:    tgt.MaxBackoff,
	}

	var lastIndex uint64
	for {
		ss, meta, err := s.ServiceMultipleTags(
			tgt.Service,
			tgt.tags,
			tgt.Healthy,
			&api.QueryOptions{
				WaitIndex:         lastIndex,
				Near:              tgt.Near,
				WaitTime:          tgt.Wait,
				Datacenter:        tgt.Dc,
				AllowStale:        tgt.AllowStale,
				RequireConsistent: tgt.RequireConsistent,
			},
		)
		if err != nil {
			grpclog.Errorf("[Consul resolver] Couldn't fetch endpoints. target={%s}; error={%v}", tgt.String(), err)
			time.Sleep(bck.Duration())
			continue
		}

		bck.Reset()

		if meta.LastIndex == lastIndex {
			continue
		}

		if meta.LastIndex < lastIndex {
			// according to https://www.consul.io/api-docs/features/blocking
			// we should reset the index if it goes backward
			lastIndex = 0
		} else {
			lastIndex = meta.LastIndex
		}

		grpclog.Infof("[Consul resolver] %d endpoints fetched in(+wait) %s for target={%s}",
			len(ss),
			meta.RequestTime,
			tgt.String(),
		)

		select {
		case out <- ss:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func populateEndpoints(
	ctx context.Context,
	clientConn resolver.ClientConn,
	input <-chan []*api.ServiceEntry,
	limit int,
	agentNodeName string,
	sortOrder string,
) {
	for {
		select {
		case in := <-input:
			if sortOrder == sortSameNodeFirst {
				sort.Sort(sameNodeFirst{
					agentNodeName: agentNodeName,
					in:            in,
				})
			}

			if sortOrder == "" || sortOrder == sortByName {
				sort.Sort(byName(in))
			}

			if limit != 0 && len(in) > limit {
				in = in[:limit]
			}

			addrs := make([]resolver.Address, 0, len(in))
			for _, s := range in {
				addrs = append(addrs, resolver.Address{
					Addr: fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port),
				})
			}

			clientConn.UpdateState(resolver.State{Addresses: addrs})
		case <-ctx.Done():
			grpclog.Info("[Consul resolver] Watch has been finished")
			return
		}
	}
}
