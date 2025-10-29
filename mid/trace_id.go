package mid

import (
	"fmt"
	"time"

	"github.com/guestin/mob"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func TraceId() echo.MiddlewareFunc {
	return middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return fmt.Sprintf("%s%s", time.Now().Format("060102"), mob.GenRandomUUID()[6:])
		},
		RequestIDHandler: func(ctx echo.Context, s string) {
			ctx.Set(CtxTraceIdKey, s)
		},
	})
}

func GetTraceId(ctx echo.Context) string {
	return fmt.Sprintf("%v", ctx.Get(CtxTraceIdKey))
}
