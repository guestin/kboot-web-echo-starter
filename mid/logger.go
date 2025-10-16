package mid

import (
	"github.com/guestin/log"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func WithTraceLogger(rootLogger log.ZapLog) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			requestId := ctx.Response().Header().Get(echo.HeaderXRequestID)
			loggerWithTraceId := rootLogger.With(log.UseFields(zap.String("traceId", requestId)))
			ctx.Set(CtxZapLoggerKey, loggerWithTraceId)
			return next(ctx)
		}
	}
}

func GetTraceLogger(ctx echo.Context) log.ZapLog {
	return ctx.Get(CtxContextKey).(log.ZapLog)
}
