package web

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/guestin/kboot"
	"github.com/guestin/kboot-web-echo-starter/kerrors"
	"github.com/guestin/kboot-web-echo-starter/mid"
	"github.com/guestin/log"
	"github.com/guestin/mob/merrors"
	"github.com/guestin/mob/mvalidate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

var _gWeb *web

type web struct {
	ctx     context.Context
	echoCtx *echo.Echo
	cfg     *Config
	unit    kboot.Unit
	logger  log.ZapLog
}

func (this *web) Init() error {
	eCtx := this.echoCtx
	eCtx.HideBanner = true
	eCtx.HidePort = false
	eCtx.DisableHTTP2 = true
	eCtx.HTTPErrorHandler = this.globalErrorHandle
	eCtx.Validator = kboot.MValidator()
	// custom context
	eCtx.Use(mid.WithContext(this.ctx))
	// trace
	eCtx.Use(mid.TraceWithConfig(mid.TraceConfig{
		TargetHeader: this.cfg.Trace.TargetHeader,
	}))
	//cors
	eCtx.Use(middleware.CORS())
	//gzip
	eCtx.Use(middleware.Gzip())
	return nil
}

func (this *web) Start() error {
	return this.echoCtx.Start(this.cfg.ListenAddress)
}

func (this *web) Shutdown() error {
	return this.echoCtx.Shutdown(this.ctx)
}

func (this *web) globalErrorHandle(err error, ctx echo.Context) {
	if err == nil {
		return
	}
	errCategory := uint8(0) // means default
	var rsp merrors.Error
	status := http.StatusOK
	switch err.(type) {
	case merrors.Error:
		errCategory = 1
		// code = 0, means no error
		errors.As(err, &rsp)
	case validator.ValidationErrors, mvalidate.ValidateError:
		errCategory = 2
		rsp = kerrors.ErrBadRequestf("Bad Request :%v", err)
	case *echo.HTTPError:
		errCategory = 3
		var he *echo.HTTPError
		errors.As(err, &he)
		status = he.Code
		rsp = kerrors.Errorf(kerrors.HttpStatus2Code(he.Code), "%s", he.Message)
	case error:
		errCategory = 4
		rsp = kerrors.InternalErrf("unexpect error :%v", err)
	default:
		ctx.Echo().DefaultHTTPErrorHandler(err, ctx)
	}
	if !ctx.Response().Committed {
		_ = ctx.JSON(status, kerrors.WrapSensitiveErr(rsp))
	}
	if rsp.GetCode() < kerrors.CodeInternalServer {
		// excepted business error
		return
	}
	this.logger.Warn("api global error handler",
		zap.String("path", ctx.Path()),
		zap.Uint8("errCategory", errCategory),
		zap.Error(err))
}
