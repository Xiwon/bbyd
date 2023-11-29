package main

import(
	"bbyd/internal/model"
	"bbyd/internal/shared/config"
	"bbyd/internal/controllers/auth"
	"bbyd/internal/controllers/server"
)

func main() {
	err := config.Create()
	if err != nil {
		panic(err)
	}

	err = model.Init(config.Configs.Database)
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
