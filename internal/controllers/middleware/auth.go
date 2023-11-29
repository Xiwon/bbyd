package middleware

import(
	"net/http"
	
	"bbyd/internal/model"
	"bbyd/internal/controllers/auth"
	contro "bbyd/internal/controllers"
	resp "bbyd/pkg/utils/response"
	
	"github.com/labstack/echo/v4"
)

func TokenVerify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*resp.ResponseContext)
		claims, err := auth.GetClaimsFromHeader(c)
		// get token from header and parse it while checking validity
		if err != nil {
			return c.BYResponse(http.StatusBadRequest, err.Error(), nil)
		}

		mod, err := model.GetUsrByName(claims.Username) // get raw data from database
		if err != nil {
			return c.BYResponse(http.StatusBadRequest, "database error", err.Error())
		}
		usr := contro.UserModelToUserProfile(mod)
		c.Set("token_usr", &usr)
		// set `token_usr` to contro.UserProfile object

		return next(c)
	}
}
