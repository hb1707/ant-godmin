package setting

var Redis struct {
	Host     string
	Port     int
	Username string
	Password string
	DB       int
}

func confRedis() {
	rd, err := Cfg.GetSection("redis")
	if err == nil {
		Redis.Host = rd.Key("REDIS_HOST").MustString("localhost")
		Redis.Port = rd.Key("REDIS_PORT").MustInt(6379)
		Redis.Username = rd.Key("REDIS_USERNAME").MustString("default")
		Redis.Password = rd.Key("REDIS_PASSWORD").MustString("")
		Redis.DB = rd.Key("REDIS_DB").MustInt(0)
	}
}
