package services

import (
	"github.com/gin-gonic/gin"

	"github.com/kadragon/apis-data-spcdeinfoservice-cloud-run/internal/proxy"
)

type ServiceSpec struct {
	MountPath    string
	BaseURL      string
	AllowedPaths []string
}

var all = []ServiceSpec{
	BidPublicInfoSpec,
	GetSecuritiesProductInfoSpec,
	KorServiceSpec,
	PubDataOpnStdSpec,
	SjFestivalSpec,
	SpcdeInfoSpec,
}

func RegisterAll(r *gin.Engine, serviceKey string) {
	for _, s := range all {
		grp := r.Group(s.MountPath)
		for _, p := range s.AllowedPaths {
			grp.GET(p, proxy.NewHandler(s.BaseURL, p, serviceKey))
		}
	}
}
