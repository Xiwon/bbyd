package main

import(
	// "fmt"
	// "gorm.io/driver/postgres"
	// "gorm.io/gorm"

	"github.com/labstack/echo/v4"
	"bbyd/db"
	"bbyd/jwt"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	db.DatabaseStart() //  database connect

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(jwt.AutoLogin) //  middleware - jwt

	registerRoutes(e)

	e.Logger.Fatal(e.Start(":11451"))
}
