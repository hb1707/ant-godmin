package setting

var Volc struct {
	KeyId     string
	SecretKey string
}

func confVolc() {
	vc, err := Cfg.GetSection("volc")
	if err == nil {
		Volc.KeyId = vc.Key("KEY_ID").String()
		Volc.SecretKey = vc.Key("SECRET_KEY").String()
	}
}
