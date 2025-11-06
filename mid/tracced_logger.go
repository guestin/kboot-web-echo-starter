package mid

import (
	"github.com/guestin/log"
	"github.com/labstack/echo/v4"
)

func WithTraceLogger(rootLogger log.ZapLog) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			requestId := GetTraceId(ctx)
			loggerWithTraceId := rootLogger.With(log.UseSubTag(log.NewFixStyleText(requestId, log.Blue, false)))
			ctx.Set(CtxZapLoggerKey, loggerWithTraceId)
			return next(ctx)
		}
	}
}

func GetTraceLogger(ctx echo.Context) log.ZapLog {
	return ctx.Get(CtxZapLoggerKey).(log.ZapLog)
}
