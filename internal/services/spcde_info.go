package services

var SpcdeInfoSpec = ServiceSpec{
	MountPath: "/SpcdeInfoService",
	BaseURL:   "https://apis.data.go.kr/B090041/openapi/service/SpcdeInfoService",
	AllowedPaths: []string{
		"/getRestDeInfo",
		"/getAnniversaryInfo",
		"/get24DivisionsInfo",
	},
}
