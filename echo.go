package web

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/guestin/kboot"
	"github.com/guestin/kboot/web/internal"
	"github.com/guestin/kboot/web/kerrors"
	"github.com/guestin/kboot/web/mid"
	"github.com/guestin/mob/merrors"
	"github.com/guestin/mob/mvalidate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func _initEcho(unit kboot.Unit, cfg *Config) (kboot.ExecFunc, error) {
	ctx := unit.GetContext()
	eCtx := echo.New()
	eCtx.HideBanner = true
	eCtx.HidePort = false
	eCtx.DisableHTTP2 = true
	eCtx.HTTPErrorHandler = globalErrorHandle
	eCtx.Validator = kboot.MValidator()
	eCtx.Use(mid.WithContext(ctx))
	//recovery
	eCtx.Use(middleware.Recover())
	// request id
	eCtx.Use(middleware.RequestID())
	//cors
	eCtx.Use(middleware.CORS())
	//gzip
	eCtx.Use(middleware.Gzip())
	//logger
	loggerOption := mid.LogNone
	if cfg.Debug {
		loggerOption = mid.LogReqHeader | mid.LogRespHeader | mid.LogJson | mid.LogForm
	}
	eCtx.Use(mid.Log(loggerOption))

	if cfg.Auth != nil && cfg.Auth.Enabled {
		eCtx.Use(mid.AuthWithConfig(*cfg.Auth))
	} else {
		eCtx.Use(mid.Auth())
	}
	// start watcher
	exitChan := make(chan error)

	return func(unit kboot.Unit) kboot.ExitResult {
		for _, opt := range _options {
			err := opt.apply(eCtx)
			if err != nil {
				internal.Log.Panic("use option failed ", err)
			}
		}
		go func() {
			exitChan <- eCtx.Start(cfg.ListenAddress)
		}()
		select {
		case err := <-exitChan:
			internal.Log.Info("API server exit", zap.Error(err))
			return kboot.NewSuccessResult()
		case <-unit.Done():
			_ = eCtx.Shutdown(unit.GetContext())
		}
		return kboot.NewSuccessResult()
	}, nil
}

func globalErrorHandle(err error, ctx echo.Context) {
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
	internal.Log.Warn("api global error handler",
		zap.String("path", ctx.Path()),
		zap.Uint8("errCategory", errCategory),
		zap.Error(err))
}
