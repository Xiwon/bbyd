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

	api := e.Group("/api/v1", mdware.UseResponseContext)
	{
		user := api.Group("/user")
		{
			user.GET("/:name", contro.UserGET, mdware.TokenVerify) // user index
			user.POST("", contro.RegisterPOST) // register
			user.PUT("/:name", contro.SetinfoPUT, mdware.TokenVerify)
			user.DELETE("/:name", contro.DeletePOST, mdware.TokenVerify) // delete
			
			user.GET("/token", contro.LoginPOST) // login
			user.DELETE("/token", contro.LogoutPOST, mdware.TokenVerify) // logout
		}
	}
}