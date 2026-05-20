package setting

var Redis struct {
	Host     string
	Port     int
	Username string
	Password string
	DB       int
}

func confRedis() {
	rd, _ := Cfg.GetSection("redis")
	Redis.Host = getString(rd, "redis", "REDIS_HOST", "localhost")
	Redis.Port = getInt(rd, "redis", "REDIS_PORT", 6379)
	Redis.Username = getString(rd, "redis", "REDIS_USERNAME", "default")
	Redis.Password = getString(rd, "redis", "REDIS_PASSWORD", "")
	Redis.DB = getInt(rd, "redis", "REDIS_DB", 0)
}
