package services

import (
	"github.com/gin-gonic/gin"
)

type ServiceSpec struct {
	MountPath    string
	BaseURL      string
	AllowedPaths []string
}

// HandlerParams carries the named inputs for constructing a proxy handler,
// preventing silent swaps between same-typed positional arguments.
type HandlerParams struct {
	BaseURL    string
	Path       string
	ServiceKey string
}

// HandlerFactory constructs a gin.HandlerFunc for a proxied upstream path.
type HandlerFactory func(HandlerParams) gin.HandlerFunc

var all = []ServiceSpec{
	BidPublicInfoSpec,
	GetSecuritiesProductInfoSpec,
	KorServiceSpec,
	PubDataOpnStdSpec,
	SjFestivalSpec,
	SpcdeInfoSpec,
}

func RegisterAll(r *gin.Engine, serviceKey string, factory HandlerFactory) {
	for _, s := range all {
		grp := r.Group(s.MountPath)
		for _, p := range s.AllowedPaths {
			grp.GET(p, factory(HandlerParams{BaseURL: s.BaseURL, Path: p, ServiceKey: serviceKey}))
		}
	}
}
