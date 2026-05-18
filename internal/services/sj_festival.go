package services

var SjFestivalSpec = ServiceSpec{
	MountPath: "/sjFestival",
	BaseURL:   "https://apis.data.go.kr/5690000/sjFestival",
	AllowedPaths: []string{
		"/sj_00000360",
	},
}
