package server

import "github.com/stockyard-dev/stockyard-rustler/internal/license"

type Limits struct {
	MaxScans int
	MaxDepth int
	SSLCheck bool
}

var freeLimits = Limits{MaxScans: 10, MaxDepth: 2, SSLCheck: true}
var proLimits = Limits{MaxScans: 0, MaxDepth: 10, SSLCheck: true}

func LimitsFor(info *license.Info) Limits {
	if info != nil && info.IsPro() { return proLimits }
	return freeLimits
}

func LimitReached(limit, current int) bool {
	if limit == 0 { return false }
	return current >= limit
}
