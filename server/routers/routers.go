package routers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/routers/html"
	"github.com/hb1707/ant-godmin/routers/json"
	"github.com/hb1707/ant-godmin/setting"
	"net/http"
)

// List 路由列表设定
func List(isRelease bool, allowOrigins []string, allowHeader []string) *gin.Engine {
	r := gin.New()
	config := cors.DefaultConfig()
	//config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AllowOrigins = allowOrigins
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	if len(allowHeader) > 0 {
		config.AddAllowHeaders(allowHeader...)
	} else {
		config.AddAllowHeaders("Authorization,x-requested-with,withcredentials")
	}
	r.Use(gin.Logger(), gin.Recovery(), cors.New(config))
	if isRelease {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode) //debug
	}
	r.Static("/upload", setting.Upload.LocalPath)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})
	m := auth.Middleware()
	api := r.Group("/api")
	{
		systemGroup := api.Group("/system")
		systemGroup.GET("/config", json.Config)
		systemGroup.GET("/qywx-connect", json.QyWxConnect)
		systemGroup.GET("/qywx-jsconfig", json.QyWxJsConfig)
		systemGroup.GET("/qywx-agent-jsconfig", json.QyWxAgentJsConfig)
		if setting.IsCMS {
			systemGroup.POST("/reg/account", json.RegisterWithPassword)
		}
		systemGroup.POST("/login/account", json.LoginWithPasswordOrQywxCode)
		systemGroup.POST("/logout", m.LogoutHandler)
		systemGroup.Use(m.MiddlewareFunc(), auth.CheckTokenUser)
		{
			systemGroup.GET("/auth/refresh", json.RefreshToken)
			systemGroup.GET("/auth/self", json.GetUser)
			systemGroup.GET("/qywx-launch-code", json.WxGetLaunchCode)
			systemGroup.POST("/file/upload/:path", json.UploadOSS)
			systemGroup.POST("/file/download/:path", json.DownloadFile)
			systemGroup.POST("/file/local-ipfs/:path", json.AddIPFS)
			systemGroup.POST("/file/local-oss/:path", json.AddOSS)
			systemGroup.POST("/file/local/:path", json.UploadLocal)
		}
		if setting.IsCMS {
			cmsGroup := api.Group("/cms")
			{
				cmsGroup.GET("/detail/:table", json.FetchOne)
				cmsGroup.GET("/list/:table", json.FetchAll)
				dataGroup := cmsGroup.Group("/data")
				dataGroup.Use(m.MiddlewareFunc(), auth.CheckTokenUser)
				{
					dataGroup.POST("/add/:table", json.Create)
					dataGroup.POST("/update/:table/:id", json.Update)
					dataGroup.DELETE("/delete/:table/:id", json.Delete)
				}
				tableGroup := cmsGroup.Group("/table")
				tableGroup.Use(m.MiddlewareFunc(), auth.CheckTokenUser)
				{
					tableGroup.GET("/list", json.FetchTablesAll)
					tableGroup.GET("/detail/:table", json.DetailTable)
					tableGroup.POST("/edit/:table", json.EditTables)
					tableGroup.DELETE("/del/:table", json.DelTables)
					tableGroup.GET("/fields/detail", json.DetailField)
					tableGroup.GET("/fields/list/:table", json.ListFields)
					tableGroup.POST("/fields/edit/:table", json.EditFields)
					tableGroup.DELETE("/fields/del/:table", json.DelFields)
				}
			}
		}
	}

	qyGroup := r.Group("/api/auth/qy")
	{
		qyGroup.POST("/wx-user", json.WxGetWorkUser)
		qyGroup.POST("/wx-reg", json.WorkRegister)
	}

	logsGroup := r.Group("/api/logs") //.Use(m.MiddlewareFunc(), auth.CheckTokenUser)
	{
		logsGroup.POST("/add", json.AddLog)
	}

	return r
}

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", html.DataHtml(c, map[string]interface{}{
		"title": "ANT GODMIN",
	}))
	if c.Request.URL.Path != "/" {
		http.Error(c.Writer, "Not found", http.StatusNotFound)
		return
	}
	if c.Request.Method != "GET" {
		http.Error(c.Writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
