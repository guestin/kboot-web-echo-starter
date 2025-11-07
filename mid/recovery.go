package mid

import (
	"github.com/guestin/kboot-web-echo-starter/kerrors"
	"github.com/guestin/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func Recovery(logger log.ZapLog) echo.MiddlewareFunc {
	if logger == nil {
		panic("Logger must not be nil")
	}
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisablePrintStack: true,
		LogErrorFunc: func(ctx echo.Context, err error, stack []byte) error {
			traceId := GetTraceId(ctx)
			if traceId != "" {
				logger = logger.With(log.UseSubTag(log.NewFixStyleText(traceId, log.Blue, false)))
			}
			logger.Error(
				"panic recovery",
				zap.Error(err),
				zap.Binary("stack", stack[:]),
			)
			return kerrors.InternalErr()
		},
	})
}
