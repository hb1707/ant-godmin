package setting

var Volc struct {
	ApiKey    string
	KeyId     string
	SecretKey string
}

func confVolc() {
	vc, _ := Cfg.GetSection("volc")
	Volc.ApiKey = GetString(vc, "volc", "API_KEY", "")
	Volc.KeyId = GetString(vc, "volc", "KEY_ID", "")
	Volc.SecretKey = GetString(vc, "volc", "KEY_SECRET", "")
}
