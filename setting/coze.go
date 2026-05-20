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
	Coze.ClientId = getString(cz, "coze", "CLIENT_ID", "")
	Coze.ClientSecret = getString(cz, "coze", "CLIENT_SECRET", "")

	czTeam, _ := Cfg.GetSection("coze_team")
	CozeTeam.ClientId = getString(czTeam, "coze_team", "CLIENT_ID", "")
	CozeTeam.ClientSecret = getString(czTeam, "coze_team", "CLIENT_SECRET", "")
	CozeTeam.EnterpriseId = getString(czTeam, "coze_team", "ENTERPRISE_ID", "")
	CozeTeam.OrganizationId = getString(czTeam, "coze_team", "ORGANIZATION_ID", "")
	CozeTeam.TermToken = getString(czTeam, "coze_team", "TERM_TOKEN", "")
	CozeTeam.AdminUserId = getString(czTeam, "coze_team", "ADMIN_USER_ID", "")
}
