package setting

var Dify struct {
	BaseUrl   string
	SecretKey string
}

func confDify() {
	vc, _ := Cfg.GetSection("dify")
	Dify.BaseUrl = GetString(vc, "dify", "BASE_URL", "")
	Dify.SecretKey = GetString(vc, "dify", "KEY_SECRET", "")
}
