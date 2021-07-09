package consul

import (
	"github.com/hashicorp/consul/api"
)

const (
	sortNone          = "none"
	sortByName        = "byName"
	sortSameNodeFirst = "sameNodeFirst"
)

// sameNodeFirst sorts services so that services on the same
// node go first, the rest services remain unchanged. Useful
// with near=_agent&limit=1 to prevent from reconnect because
// of random network issues.
type sameNodeFirst struct {
	agentNodeName string
	in            []*api.ServiceEntry
}

func (n sameNodeFirst) Len() int      { return len(n.in) }
func (n sameNodeFirst) Swap(i, j int) { n.in[i], n.in[j] = n.in[j], n.in[i] }

func (n sameNodeFirst) Less(i, j int) bool {
	if n.in[i].Node.Node == n.agentNodeName && n.in[j].Node.Node != n.agentNodeName {
		return true
	}

	return false
}

// byName sorts services by address lexicographic order.
type byName []*api.ServiceEntry

func (p byName) Len() int           { return len(p) }
func (p byName) Less(i, j int) bool { return p[i].Service.Address < p[j].Service.Address }
func (p byName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
