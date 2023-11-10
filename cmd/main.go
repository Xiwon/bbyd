package main

import(
	"bbyd/internal/model"
	"bbyd/internal/shared/config"
	"bbyd/internal/controllers/auth"
	"bbyd/internal/controllers/server"
)

func main() {
	conf, err := config.Create()
	if err != nil {
		panic(err)
	}

	err = model.Init(conf.Database)
	if err != nil {
		panic(err)
	}

	err = auth.Init(conf.Authorization)
	if err != nil {
		panic(err)
	}

	err = server.Run(conf.Server)
	if err != nil {
		panic(err)
	}
}
