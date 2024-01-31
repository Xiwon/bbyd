package middleware

import(
	"bbyd/pkg/utils/response"
	"github.com/labstack/echo/v4"
)

func UseResponseContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// use custom context with custom response func defined
		return next(&response.ResponseContext{c})
	}
}