package mid

import (
	"github.com/guestin/kboot-web-echo-starter/internal"
	"github.com/labstack/echo/v4"
)

func ReqBodyReplay() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			if req.Body != nil && req.ContentLength > 0 {
				c.Request().Body = internal.NewReplayBuffer(c.Request().Body)
			}
			return next(c)
		}
	}
}
