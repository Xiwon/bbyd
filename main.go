package main

import(
	"bbyd/model"
	"bbyd/shared/config"
	"bbyd/controllers/auth"
	"bbyd/controllers/server"
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
