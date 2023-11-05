package server

import(
	"strconv"

	"bbyd/pkg/utils/logs"
	"bbyd/internal/shared/config"
	contro "bbyd/internal/controllers"
	mdware "bbyd/internal/controllers/middleware"

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

	api := e.Group("/api/v1")
	{
		user := api.Group("/user")
		{
			user.GET("", contro.UserGET, mdware.TokenVerify)
			user.GET("/", contro.UserGET, mdware.TokenVerify)
			
			user.POST("/login", contro.LoginPOST)
			user.POST("/register", contro.RegisterPOST)
			user.POST("/logout", contro.LogoutPOST, mdware.TokenVerify)

			user.POST("/setauth", contro.SetauthPOST, mdware.TokenVerify)
			user.POST("/setinfo", contro.SetinfoPOST, mdware.TokenVerify)
			user.POST("/delete", contro.DeletePOST, mdware.TokenVerify)
		}
	}
}