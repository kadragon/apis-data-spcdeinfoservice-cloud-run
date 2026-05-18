package services

var GetSecuritiesProductInfoSpec = ServiceSpec{
	MountPath: "/GetSecuritiesProductInfoService",
	BaseURL:   "https://apis.data.go.kr/1160100/service/GetSecuritiesProductInfoService",
	AllowedPaths: []string{
		"/getETFPriceInfo",
		"/getETNPriceInfo",
		"/getELWPriceInfo",
	},
}
