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
	AdminUserId    string // 管理员用户ID
	TermToken      string
}

func confCoze() {
	cz, _ := Cfg.GetSection("coze")
	Coze.ClientId = GetString(cz, "coze", "CLIENT_ID", "")
	Coze.ClientSecret = GetString(cz, "coze", "CLIENT_SECRET", "")

	czTeam, _ := Cfg.GetSection("coze_team")
	CozeTeam.ClientId = GetString(czTeam, "coze_team", "CLIENT_ID", "")
	CozeTeam.ClientSecret = GetString(czTeam, "coze_team", "CLIENT_SECRET", "")
	CozeTeam.EnterpriseId = GetString(czTeam, "coze_team", "ENTERPRISE_ID", "")
	CozeTeam.OrganizationId = GetString(czTeam, "coze_team", "ORGANIZATION_ID", "")
	CozeTeam.TermToken = GetString(czTeam, "coze_team", "TERM_TOKEN", "")
	CozeTeam.AdminUserId = GetString(czTeam, "coze_team", "ADMIN_USER_ID", "")
}
