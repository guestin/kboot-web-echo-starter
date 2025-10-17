package mid

import (
	"bytes"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/guestin/kboot"
	"github.com/guestin/kboot-web-echo-starter/kerrors"
	"github.com/guestin/mob/merrors"
	"github.com/labstack/echo/v4"
	"github.com/ooopSnake/assert.go"
	"go.uber.org/zap"

	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type (
	// WrapConfig defines the config for Format middleware.
	WrapConfig struct {
		AllowDuplicateBind bool
		SkipFormat         bool
	}
	WrapOption interface {
		apply(cfg *WrapConfig)
	}
	wrapOptionFunc func(cfg *WrapConfig)
)

func (f wrapOptionFunc) apply(cfg *WrapConfig) {
	f(cfg)
}

func WrapAllowDuplicateBind() WrapOption {
	return wrapOptionFunc(func(cfg *WrapConfig) {
		cfg.AllowDuplicateBind = true
	})
}

func SkipFormat() WrapOption {
	return wrapOptionFunc(func(cfg *WrapConfig) {
		cfg.SkipFormat = true
	})
}

func Wrap(handler interface{}, option ...WrapOption) echo.HandlerFunc {
	cfg := &WrapConfig{}
	for _, opt := range option {
		if opt != nil {
			opt.apply(cfg)
		}
	}
	handlerValue, ok := handler.(reflect.Value)
	if !ok {
		handlerValue = reflect.ValueOf(handler)
	}
	assert.Must(handlerValue.Kind() == reflect.Func, "handler must be a func !").Panic()
	handlerType := handlerValue.Type()
	fName := getFuncName(handlerValue)
	inType, inFlags := checkInParam(fName, handlerType)
	_, outFlags := checkOutParam(fName, handlerType)
	reqIsPtr := false
	reqIsSlice := false
	if inFlags&handlerHasReqData != 0 {
		if inType.Kind() == reflect.Ptr {
			inType = inType.Elem()
			reqIsPtr = true
		} else if inType.Kind() == reflect.Slice {
			reqIsSlice = true
		}
	}
	return func(ctx echo.Context) error {
		defer func() {
			err := recover()
			if err != nil {
				kboot.GetTaggedZapLogger("mid").Error("panic recover",
					zap.Any("error", err),
					zap.Any("stack", string(debug.Stack())),
				)
				ctx.Set(CtxStatusKey, http.StatusInternalServerError)
				ctx.Set(CtxErrorKey, kerrors.InternalErr("Server is busy"))
			}
			checkErrAndFlush(ctx, cfg)
		}()
		inParams := make([]reflect.Value, 0)
		if inFlags&handlerHasCtx != 0 {
			inParams = append(inParams, reflect.ValueOf(ctx))
		}
		//has req data
		if inFlags&handlerHasReqData != 0 {
			var err error
			var req interface{}
			if reqIsSlice {
				req = reflect.New(reflect.SliceOf(inType.Elem())).Interface()
			} else {
				req = reflect.New(inType).Interface()
			}
			if cfg.AllowDuplicateBind {
				reqBody := make([]byte, 0)
				if ctx.Request().Body != nil {
					reqBody, _ = io.ReadAll(ctx.Request().Body)
				}
				ctx.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody))
				// bind
				err = ctx.Bind(req)
				// reset the body for next bind
				ctx.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody))
				//stream := internal.NewReplayBuffer(ctx.Request().Body)
				//ctx.Request().Body = stream
				// bind
				//err = ctx.Bind(req)
				//// seek back to start
				//_, _ = stream.Seek(0, io.SeekStart)
			} else {
				// bind
				err = ctx.Bind(req)
			}
			if err != nil {
				ctx.Set(CtxStatusKey, http.StatusOK)
				ctx.Set(CtxErrorKey, kerrors.ErrBadRequestf("Bad Request :%s ", err))
				return nil
			}
			if !reqIsPtr {
				req = reflect.ValueOf(req).Elem().Interface()
			}
			//validate
			err = ctx.Validate(req)
			if err != nil {
				ctx.Set(CtxStatusKey, http.StatusOK)
				ctx.Set(CtxErrorKey, kerrors.ErrBadRequestf("Bad Request :%s ", err))
				return nil
			}
			inParams = append(inParams, reflect.ValueOf(req))
		}
		//invoke
		outs := handlerValue.Call(inParams)
		rspErrIdx := -1
		rspDataIdx := -1
		//has rsp data
		if outFlags&handlerHasRsp != 0 {
			rspErrIdx = 1
			rspDataIdx = 0
		} else {
			rspErrIdx = 0
		}
		var err error
		if !outs[rspErrIdx].IsNil() { // check error
			err = outs[rspErrIdx].Interface().(error)
			ctx.Set(CtxErrorKey, err)
		}
		if rspDataIdx != -1 {
			oKind := outs[rspDataIdx].Kind()
			if oKind == reflect.Ptr || oKind == reflect.Struct {
				if !(outs[rspDataIdx]).IsNil() {
					rsp := outs[rspDataIdx].Interface()
					ctx.Set(CtxRespKey, rsp)
				}
			} else {
				ctx.Set(CtxRespKey, outs[rspDataIdx].Interface())
			}
		}
		return nil
	}
}

func getFuncName(fv reflect.Value) string {
	fnName := runtime.FuncForPC(reflect.Indirect(fv).Pointer()).Name()
	idx := strings.LastIndex(fnName, "/")
	if idx != -1 {
		fnName = fnName[idx+1:]
	}
	idx = strings.LastIndex(fnName, "-")
	if idx != -1 {
		fnName = fnName[:idx]
	}
	return fnName
}

const (
	handlerHasCtx uint32 = 1 << iota
	handlerHasReqData
	handlerHasRsp
)

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()
var typeOfContext = reflect.TypeOf((*echo.Context)(nil)).Elem()

func checkInParam(name string, t reflect.Type) (reflect.Type, uint32) {
	var handlerFlags uint32 = 0
	var inParamType reflect.Type = nil
	inNum := t.NumIn()
	assert.Mustf(inNum >= 0 && inNum <= 2,
		"'%s' not valid : inNum len must be 0() or 1(Any) or 2(echo.Context,Any)", name).Panic()
	//
	switch inNum {
	case 0:
		//func()
	case 1:
		// func foo(Context)
		if t.In(0) == typeOfContext {
			handlerFlags = handlerFlags | handlerHasCtx
		} else {
			// func foo(param1)
			handlerFlags = handlerFlags | handlerHasReqData
			in1 := t.In(0)
			inParamType = in1
		}
	case 2:
		// func foo(Context,param1)
		in0 := t.In(0)
		assert.Mustf(in0 == typeOfContext,
			"'%s' not valid :first in param must be echo.Context", name).Panic()
		in1 := t.In(1)
		inParamType = in1
		handlerFlags = handlerFlags | handlerHasCtx | handlerHasReqData
	default:
		assert.Mustf(false, "'%s' not valid :illegal func in params num", name).Panic()
	}
	return inParamType, handlerFlags
}

func checkOutParam(name string, t reflect.Type) (reflect.Type, uint32) {
	var handlerFlags uint32 = 0
	var outParamType reflect.Type = nil
	outNum := t.NumOut()
	assert.Mustf(outNum > 0 && outNum <= 2,
		"'%s' not valid :outNum len must be 1(error) or 2(any,error)", name).Panic()
	lastOut := t.Out(outNum - 1)
	assert.Mustf(lastOut == typeOfError,
		"'%s' not valid :the last out param must be 'error'", name).Panic()
	switch outNum {
	case 1:
		//fun(xxx)error
	case 2:
		outParamType = t.Out(0)
		handlerFlags = handlerFlags | handlerHasRsp
	default:
		assert.Mustf(false, "'%s' not valid :illegal func return params num", name).Panic()
	}
	return outParamType, handlerFlags
}

func checkErrAndFlush(ctx echo.Context, config *WrapConfig) {
	//requestId := ctx.Request().Header.Get(echo.HeaderXRequestID)
	statusCode := 200
	statusI := ctx.Get(CtxStatusKey)
	if statusI != nil {
		if status, ok := statusI.(int); ok && status > 0 {
			statusCode = status
		}
	}
	var resp interface{} = nil
	ctxErr := ctx.Get(CtxErrorKey)
	if ctxErr != nil {
		if config.SkipFormat {
			resp = ctxErr
			goto flush
		}
		switch ctxErr.(type) {
		case merrors.Error:
			resp = ctxErr
		case *echo.HTTPError:
			he := ctxErr.(*echo.HTTPError)
			resp = merrors.Errorf0(he.Code, "%s", fmt.Sprint(he.Message))
		default:
			resp = kerrors.ErrInternalf("Server is busy : %s", ctxErr)
		}
		goto flush
	}
	resp = ctx.Get(CtxRespKey)
	if config.SkipFormat {
		goto flush
	}
	if resp != nil {
		switch resp.(type) {
		case merrors.Error:
			goto flush
		default:
			resp = kerrors.OkResult(resp)
		}
	} else {
		resp = kerrors.OkResult(nil)
	}

flush:
	// has rsp & no error need write response,otherwise err handler will handle
	if !ctx.Response().Committed && resp != nil {
		_ = ctx.JSON(statusCode, resp)
		//_ = ctx.JSONPretty(statusCode, resp,jsonIndent)
	}
}
