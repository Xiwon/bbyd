package response

import(
	"github.com/labstack/echo/v4"
)

type MsgResponse struct {
	Code int
	Status string
	Message string
	Data interface{}
}

type ResponseContext struct {
	echo.Context
}

func (r *ResponseContext) BYResponse(
	code int, msg string, data interface{}) error {
	
	judge := "success"
	if code >= 400 {
		judge = "failed"
	}

	return r.JSON(code, &MsgResponse{
		Code: code,
		Status: "[" + r.Request().URL.Path + "] " + judge,
		Message: msg,
		Data: data,
	})
}