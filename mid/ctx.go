package mid

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
)

func WithContext(ctx context.Context) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(CtxContextKey, ctx)
			return next(c)
		}
	}
}

//goland:noinspection ALL
func GetContext(eCtx echo.Context) context.Context {
	return eCtx.Get(CtxContextKey).(context.Context)
}

func UnwrapContext(e echo.Context) context.Context {
	if ctx, ok := e.Get(CtxContextKey).(context.Context); ok {
		return ctx
	}
	panic("no context available")
}

//goland:noinspection ALL
func UnwrapTimeoutContext(e echo.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(UnwrapContext(e), timeout)
}
