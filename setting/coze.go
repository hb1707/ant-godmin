package setting

var Coze struct {
	ClientId     string
	ClientSecret string
}
var CozeTeam struct {
	ClientId       string
	ClientSecret   string
	EnterpriseId   string // 企业ID
	OrganizationId string // 组织ID
}

func confCoze() {
	cz, err := Cfg.GetSection("coze")
	if err == nil {
		Coze.ClientId = cz.Key("CLIENT_ID").String()
		Coze.ClientSecret = cz.Key("CLIENT_SECRET").String()
	}
	czTeam, err := Cfg.GetSection("coze_team")
	if err == nil {
		CozeTeam.ClientId = czTeam.Key("CLIENT_ID").String()
		CozeTeam.ClientSecret = czTeam.Key("CLIENT_SECRET").String()
		CozeTeam.EnterpriseId = czTeam.Key("ENTERPRISE_ID").String()
		CozeTeam.OrganizationId = czTeam.Key("ORGANIZATION_ID").String()
	}
}
