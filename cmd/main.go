package main

import(
	"os"

	"bbyd/internal/model"
	"bbyd/internal/shared/config"
	"bbyd/internal/controllers/auth"
	"bbyd/internal/controllers/server"
)

func main() {
	args := os.Args
	var configTomlPath = ""
	if len(args) < 2 {
		configTomlPath = "../config/config.toml"
	} else {
		configTomlPath = args[1]
	}

	err := config.Create(configTomlPath)
	if err != nil {
		panic(err)
	}

	err = model.Init(config.Configs.Database, config.Configs.RedisConfig)
	if err != nil {
		panic(err)
	}

	err = auth.Init(config.Configs.Authorization)
	if err != nil {
		panic(err)
	}

	err = server.Run(config.Configs.Server)
	if err != nil {
		panic(err)
	}
}
