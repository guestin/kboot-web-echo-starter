package mid

import (
	"fmt"
	"time"

	"github.com/guestin/mob"
	"github.com/labstack/echo/v4"
)

// RequestIDConfig defines the config for RequestID middleware.
type RequestIDConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Generator defines a function to generate an ID.
	// Optional. Defaults to generator for random string of length 32.
	Generator func() string

	// RequestIDHandler defines a function which is executed for a request id.
	RequestIDHandler func(echo.Context, string)

	// TargetHeader defines what header to look for to populate the id
	TargetHeader string
}

// DefaultRequestIDConfig is the default RequestID middleware config.
var DefaultRequestIDConfig = RequestIDConfig{
	Skipper:      DefaultSkipper,
	Generator:    generator,
	TargetHeader: echo.HeaderXRequestID,
}

// RequestID returns a X-Request-ID middleware.
func RequestID() echo.MiddlewareFunc {
	return RequestIDWithConfig(DefaultRequestIDConfig)
}

// RequestIDWithConfig returns a X-Request-ID middleware with config.
func RequestIDWithConfig(config RequestIDConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultRequestIDConfig.Skipper
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
			rid := req.Header.Get(config.TargetHeader)
			if rid == "" {
				rid = config.Generator()
			}
			defer func() {
				ctx.Set(CtxTraceIdKey, nil)
			}()
			ctx.Set(CtxTraceIdKey, rid)
			res.Header().Set(config.TargetHeader, rid)
			if config.RequestIDHandler != nil {
				config.RequestIDHandler(ctx, rid)
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
