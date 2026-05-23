package services

import (
	"github.com/gin-gonic/gin"
)

type ServiceSpec struct {
	MountPath    string
	BaseURL      string
	AllowedPaths []string
}

// HandlerFactory constructs a gin.HandlerFunc for a proxied upstream path.
type HandlerFactory func(baseURL, path, serviceKey string) gin.HandlerFunc

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
			grp.GET(p, factory(s.BaseURL, p, serviceKey))
		}
	}
}
