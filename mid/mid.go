package mid

import "github.com/labstack/echo/v4"

type (
	// Skipper defines a function to skip middleware.
	Skipper func(echo.Context) bool
)

// BeforeFunc defines a function which is executed just before the middleware.
type BeforeFunc func(c echo.Context) error

// DefaultSkipper returns false which processes the middleware.
//
//goland:noinspection ALL
func DefaultSkipper(echo.Context) bool {
	return false
}
