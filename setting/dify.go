package setting

var Dify struct {
	BaseUrl   string
	SecretKey string
}

func confDify() {
	vc, err := Cfg.GetSection("dify")
	if err == nil {
		Dify.BaseUrl = vc.Key("BASE_URL").String()
		Dify.SecretKey = vc.Key("KEY_SECRET").String()
	}
}
