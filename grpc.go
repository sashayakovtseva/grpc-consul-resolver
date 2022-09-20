package consul

import (
	"context"
	"fmt"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

//go:generate mockgen -package=consul -destination=client_conn_mock_test.go google.golang.org/grpc/resolver ClientConn

// init function for resolver registration.
func init() {
	resolver.Register(&builder{})
}

type grpcResolver struct {
	r      *Resolver
	cancel context.CancelFunc
}

// ResolveNow does nothing.
func (r grpcResolver) ResolveNow(resolver.ResolveNowOptions) {}

// Close stops underlying goroutines and releases the resources.
func (r grpcResolver) Close() {
	r.cancel()
}

// builder implements resolver.Builder and is used for constructing all consul resolvers.
type builder struct{}

func (b *builder) Build(
	target resolver.Target,
	cc resolver.ClientConn,
	_ resolver.BuildOptions,
) (resolver.Resolver, error) {
	r, err := NewResolver(target.URL.String(), WithLogger(grpcGlobalLogger{}))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	pipe := r.WatchServiceChanges(ctx)

	go populateEndpoints(ctx, cc, pipe)

	return &grpcResolver{cancel: cancel}, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (b *builder) Scheme() string {
	return schemeName
}

func populateEndpoints(ctx context.Context, clientConn resolver.ClientConn, input <-chan []*api.ServiceEntry) {
	for {
		select {
		case in := <-input:
			addrs := make([]resolver.Address, 0, len(in))
			for _, s := range in {
				addrs = append(addrs, resolver.Address{
					Addr: fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port),
				})
			}

			if err := clientConn.UpdateState(resolver.State{Addresses: addrs}); err != nil {
				grpclog.Errorf("failed to update connection stats: %v", err)
			}
		case <-ctx.Done():
			grpclog.Info("[Consul resolver] Watch has been finished")
			return
		}
	}
}
