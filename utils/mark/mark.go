package mark

import(
	"net/http"
)

const (
	OK int = http.StatusOK
	Forbdn int = http.StatusForbidden
	BadRqst int = http.StatusBadRequest
	UnAuth int = http.StatusUnauthorized
)