package server

import (
	"strconv"

	contro "bbyd/internal/controllers"
	mdware "bbyd/internal/controllers/middleware"
	"bbyd/internal/shared/config"
	"bbyd/pkg/utils/logs"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func Run(s config.Server) error {
	e := echo.New()
	routes(e)
	adr := ":" + strconv.Itoa(s.Port)
	err := e.Start(adr)
	if err != nil {
		logs.Error("server start failed at "+adr+". ", zap.Error(err))
		return err
	}
	return nil
}

func routes(e *echo.Echo) {
	e.Use(echomw.Logger())

	path := config.Configs.Constants.ApiPathRoot
	api := e.Group(path, mdware.UseResponseContext)
	{
		user := api.Group("/user")
		{
			user.GET("/:name", contro.UserIndexHandler, mdware.TokenVerify) // user index
			user.POST("", contro.RegisterHandler)                           // register
			user.PUT("/:name", contro.SetinfoHandler, mdware.TokenVerify)   // change user info
			user.DELETE("/:name", contro.DeleteHandler, mdware.TokenVerify) // delete

			user.GET("/token", contro.LoginHandler)                         // login
			user.DELETE("/token", contro.LogoutHandler, mdware.TokenVerify) // logout
		}
	}
}
