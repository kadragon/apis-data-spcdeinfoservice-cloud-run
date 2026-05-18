package services

var PubDataOpnStdSpec = ServiceSpec{
	MountPath: "/PubDataOpnStdService",
	BaseURL:   "https://apis.data.go.kr/1230000/ao/PubDataOpnStdService",
	AllowedPaths: []string{
		"/getDataSetOpnStdBidPblancInfo",
		"/getDataSetOpnStdScsbidInfo",
		"/getDataSetOpnStdCntrctInfo",
	},
}
