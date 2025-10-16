package web

import (
	"context"
	"fmt"
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
	options []Option
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
	//recovery
	eCtx.Use(middleware.Recover())
	// request id
	eCtx.Use(middleware.RequestID())
	//cors
	eCtx.Use(middleware.CORS())
	//gzip
	eCtx.Use(middleware.Gzip())
	// trace zap logger
	eCtx.Use(mid.WithTraceLogger(this.logger))
	//logger
	loggerOption := mid.LogNone
	if this.cfg.Debug {
		loggerOption = mid.LogReqHeader | mid.LogRespHeader | mid.LogJson | mid.LogForm
	}
	eCtx.Use(mid.Dump(loggerOption))

	if this.cfg.Auth != nil {
		eCtx.Use(mid.AuthWithConfig(*this.cfg.Auth))
	} else {
		eCtx.Use(mid.Auth())
	}
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
	switch err.(type) {
	case merrors.Error:
		errCategory = 1
		_ = ctx.JSON(http.StatusOK, err)
		// code = 0, means no error
		if !ctx.Response().Committed {
			_ = ctx.JSON(http.StatusOK, err)
		}
		if err.(merrors.Error).GetCode() == 0 {
			return
		}
	case validator.ValidationErrors, mvalidate.ValidateError:
		errCategory = 2
		_ = ctx.JSON(http.StatusOK,
			merrors.ErrorWrap0(err, kerrors.CodeBadRequest, "bad request params"))
	case *echo.HTTPError:
		errCategory = 3
		he := err.(*echo.HTTPError)
		_ = ctx.JSON(http.StatusOK,
			merrors.Errorf0(kerrors.HttpStatus2Code(he.Code), "%s", fmt.Sprint(he.Message)))
	case error:
		errCategory = 4
		_ = ctx.JSON(http.StatusOK,
			merrors.ErrorWrap0(err, kerrors.CodeInternalServer, "unexpect error"))
	default:
		ctx.Echo().DefaultHTTPErrorHandler(err, ctx)
	}
	this.logger.Warn("api global error handler",
		zap.String("path", ctx.Path()),
		zap.Uint8("errCategory", errCategory),
		zap.Error(err))
}
