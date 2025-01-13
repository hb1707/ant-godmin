package setting

var Coze struct {
	ClientId     string
	ClientSecret string
}

func init() {
	confCoze()
}

func confCoze() {
	cz, err := Cfg.GetSection("coze")
	if err == nil {
		Coze.ClientId = cz.Key("CLIENT_ID").String()
		Coze.ClientSecret = cz.Key("CLIENT_SECRET").String()
	}
}
