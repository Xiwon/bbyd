package response

import(
	"github.com/labstack/echo/v4"
)

type Handler struct {
	c echo.Context
	route string
}
func New(c echo.Context, route string) *Handler {
	return &Handler{c, route}
}

type MsgResponse struct {
	Code int
	Status string
	Message string
	Data interface{}
}

func (r *Handler) Re(code int, msg string, data interface{}) error {
	judge := "success"
	if code >= 400 {
		judge = "failed"
	}

	return r.c.JSONPretty(code, &MsgResponse{
		Code: code,
		Status: "[" + r.route + "] " + judge,
		Message: msg,
		Data: data,
	}, "  ")
}