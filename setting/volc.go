package setting

var Volc struct {
	ApiKey    string
	KeyId     string
	SecretKey string
}

func confVolc() {
	vc, _ := Cfg.GetSection("volc")
	Volc.ApiKey = getString(vc, "volc", "API_KEY", "")
	Volc.KeyId = getString(vc, "volc", "KEY_ID", "")
	Volc.SecretKey = getString(vc, "volc", "KEY_SECRET", "")
}
