package server

import(
	"strconv"

	"bbyd/utils/logs"
	"bbyd/shared/config"
	contro "bbyd/controllers"
	mdware "bbyd/controllers/middleware"

	"go.uber.org/zap"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func Run(s config.Server) error {
	e := echo.New()
	routes(e)
	adr := ":" + strconv.Itoa(s.Port)
	err := e.Start(adr)
	if err != nil {
		logs.Error("server start failed at " + adr + ". ", zap.Error(err))
		return err
	}
	return nil
}

func routes(e *echo.Echo) {
	e.Use(echomw.Logger())

	e.GET("/user", contro.UserGET, mdware.TokenVerify)
	
	e.GET("/login", contro.LoginPOST)
	e.GET("/register", contro.RegisterPOST)
	e.POST("/logout", contro.LogoutPOST, mdware.TokenVerify)

	e.GET("/setauth", contro.SetauthPOST, mdware.TokenVerify)
	e.GET("/setinfo", contro.SetinfoPOST, mdware.TokenVerify)
	e.POST("/delete", contro.DeletePOST, mdware.TokenVerify)
}