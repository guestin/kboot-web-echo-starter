package mid

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
)

const CtxKey = "guestin.web.ctx"

func WithContext(ctx context.Context) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(CtxKey, ctx)
			return next(c)
		}
	}
}

//goland:noinspection ALL
func GetContext(eCtx echo.Context) context.Context {
	return eCtx.Get(CtxKey).(context.Context)
}

func UnwrapContext(e echo.Context) context.Context {
	if ctx, ok := e.Get(CtxKey).(context.Context); ok {
		return ctx
	}
	panic("no context available")
}

//goland:noinspection ALL
func UnwrapTimeoutContext(e echo.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(UnwrapContext(e), timeout)
}
