package setting

var Dify struct {
	BaseUrl   string
	SecretKey string
}

func confDify() {
	vc, _ := Cfg.GetSection("dify")
	Dify.BaseUrl = getString(vc, "dify", "BASE_URL", "")
	Dify.SecretKey = getString(vc, "dify", "KEY_SECRET", "")
}
