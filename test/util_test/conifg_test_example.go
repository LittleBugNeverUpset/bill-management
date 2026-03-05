package util_test

import (
	"bill-management/pkg/config"
)

func config_test() {
	config.InitConfig("config")
	cfg := config.GetConfig()
	println("Server Host:", cfg.Server.Host)
	println("Server Port:", cfg.Server.Port)
	println("Server Mode:", cfg.Server.Mode)
	println("Database Host:", cfg.Database.Psql.Host)
	println("Database Port:", cfg.Database.Psql.Port)
	println("Database User:", cfg.Database.Psql.Username)
	println("Database Name:", cfg.Database.Psql.Database)
}
