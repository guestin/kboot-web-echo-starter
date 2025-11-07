package mid

import (
	"fmt"
	"time"

	"github.com/guestin/log"
	"github.com/guestin/mob"
	"github.com/labstack/echo/v4"
)

// TraceConfig defines the config for Trace middleware.
type (
	TraceConfig struct {
		Logger log.ZapLog
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Generator defines a function to generate an ID.
		// Optional. Defaults to generator for random string of length 32.
		Generator func() string

		// TraceIDHandler defines a function which is executed for a request id.
		TraceIDHandler func(echo.Context, string)

		// TargetHeader defines what header to look for to populate the id
		TargetHeader string `toml:"targetHeader" json:"targetHeader"`
	}
)

// DefaultTraceConfig is the default Trace middleware config.
var DefaultTraceConfig = TraceConfig{
	Logger:       nil,
	Skipper:      DefaultSkipper,
	Generator:    generator,
	TargetHeader: echo.HeaderXRequestID,
}

// Trace returns a X-Request-ID middleware.
func Trace(logger log.ZapLog) echo.MiddlewareFunc {
	return TraceWithConfig(TraceConfig{
		Logger: logger,
	})
}

// TraceWithConfig returns a X-Request-ID middleware with config.
func TraceWithConfig(config TraceConfig) echo.MiddlewareFunc {
	if config.Logger == nil {
		panic("Logger must not be nil")
	}
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTraceConfig.Skipper
	}
	if config.Generator == nil {
		config.Generator = generator
	}
	if config.TargetHeader == "" {
		config.TargetHeader = echo.HeaderXRequestID
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}
			req := ctx.Request()
			res := ctx.Response()
			traceId := req.Header.Get(config.TargetHeader)
			if traceId == "" {
				traceId = config.Generator()
			}
			traceLogger := config.Logger.With(log.UseSubTag(log.NewFixStyleText(traceId, log.Blue, false)))
			ctx.Set(CtxTraceIdKey, traceId)
			ctx.Set(CtxZapLoggerKey, traceLogger)
			defer func() {
				ctx.Set(CtxTraceIdKey, nil)
			}()
			res.Header().Set(config.TargetHeader, traceId)
			if config.TraceIDHandler != nil {
				config.TraceIDHandler(ctx, traceId)
			}
			return next(ctx)
		}
	}
}

func generator() string {
	return fmt.Sprintf("%s%s", time.Now().Format("060102"), mob.GenRandomUUID()[6:])
}

func GetTraceId(ctx echo.Context) string {
	return fmt.Sprintf("%v", ctx.Get(CtxTraceIdKey))
}
func GetTraceLogger(ctx echo.Context) log.ZapLog {
	return ctx.Get(CtxZapLoggerKey).(log.ZapLog)
}
