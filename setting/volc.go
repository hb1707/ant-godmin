package setting

var Volc struct {
	ApiKey    string
	KeyId     string
	SecretKey string
}

func confVolc() {
	vc, err := Cfg.GetSection("volc")
	if err == nil {
		Volc.ApiKey = vc.Key("API_KEY").String()
		Volc.KeyId = vc.Key("KEY_ID").String()
		Volc.SecretKey = vc.Key("KEY_SECRET").String()
	}
}
