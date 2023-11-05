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

	err = model.Init(conf.Db)
	if err != nil {
		panic(err)
	}

	err = auth.Init(conf.Au)
	if err != nil {
		panic(err)
	}

	err = server.Run(conf.Se)
	if err != nil {
		panic(err)
	}
}
