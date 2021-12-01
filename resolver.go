package consul

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/jpillora/backoff"
)

// Resolver is used to fetch service addressed from consul and watch for any changes.
// For compatibility reasons it optionally supports grpc logging via WithLoggerV2 option.
type Resolver struct {
	logger Logger

	t             *target
	c             consul
	agentNodeName string
}

func NewResolver(dsn string, opts ...Option) (*Resolver, error) {
	t, err := newTarget(dsn)
	if err != nil {
		return nil, err
	}

	consulClient, err := api.NewClient(t.consulConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Consul API: %w", err)
	}

	agentNodeName, err := consulClient.Agent().NodeName()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent node name: %w", err)
	}

	r := &Resolver{
		t:             t,
		c:             consulClient.Health(),
		agentNodeName: agentNodeName,
		logger:        noopLogger{},
	}

	for _, o := range opts {
		o(r)
	}

	return r, nil
}

//go:generate mockgen -source=resolver.go -package=consul -mock_names=consul=MockConsul -destination=consul_mock_test.go

// consul is introduced for tests only.
type consul interface {
	ServiceMultipleTags(
		service string,
		tags []string,
		passingOnly bool,
		q *api.QueryOptions,
	) ([]*api.ServiceEntry, *api.QueryMeta, error)
}

// WatchServiceChanges will send service addresses into the
// returned channel until passed context is cancelled.
func (r *Resolver) WatchServiceChanges(ctx context.Context) <-chan []*api.ServiceEntry {
	out := make(chan []*api.ServiceEntry, 1)

	go func() {
		defer close(out)

		bck := &backoff.Backoff{
			Factor: 2,
			Jitter: true,
			Min:    10 * time.Millisecond,
			Max:    r.t.MaxBackoff,
		}

		var lastIndex uint64
		for {
			endpoints, meta, err := r.c.ServiceMultipleTags(
				r.t.Service,
				r.t.tags,
				r.t.Healthy,
				&api.QueryOptions{
					WaitIndex:         lastIndex,
					Near:              r.t.Near,
					WaitTime:          r.t.Wait,
					Datacenter:        r.t.Dc,
					AllowStale:        r.t.AllowStale,
					RequireConsistent: r.t.RequireConsistent,
				},
			)
			if err != nil {
				r.logger.Errorf("[Consul resolver] Couldn't fetch endpoints. target={%s}; error={%v}", r.t.String(), err)
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

			r.logger.Infof("[Consul resolver] %d endpoints fetched in(+wait) %s for target={%s}",
				len(endpoints),
				meta.RequestTime,
				r.t.String(),
			)

			if r.t.Sort == sortSameNodeFirst {
				sort.Sort(sameNodeFirst{
					agentNodeName: r.agentNodeName,
					in:            endpoints,
				})
			}

			if r.t.Sort == "" || r.t.Sort == sortByName {
				sort.Sort(byName(endpoints))
			}

			if r.t.Limit != 0 && len(endpoints) > r.t.Limit {
				endpoints = endpoints[:r.t.Limit]
			}

			select {
			case out <- endpoints:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}
