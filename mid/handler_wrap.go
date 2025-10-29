package mid

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/guestin/kboot-web-echo-starter/kerrors"
	"github.com/guestin/mob/merrors"
	"github.com/labstack/echo/v4"
	"github.com/ooopSnake/assert.go"
)

type (
	// wrapCtx defines the config for Format middleware.
	wrapCtx struct {
		AllowDuplicateBind bool
		SkipFormat         bool
	}
	WrapOption interface {
		apply(cfg *wrapCtx)
	}
	wrapOptionFunc func(cfg *wrapCtx)
)

func (f wrapOptionFunc) apply(cfg *wrapCtx) {
	f(cfg)
}

func WrapAllowDuplicateBind() WrapOption {
	return wrapOptionFunc(func(cfg *wrapCtx) {
		cfg.AllowDuplicateBind = true
	})
}

func SkipFormat() WrapOption {
	return wrapOptionFunc(func(cfg *wrapCtx) {
		cfg.SkipFormat = true
	})
}

func Wrap(handler interface{}, option ...WrapOption) echo.HandlerFunc {
	cfg := &wrapCtx{}
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
	return func(ctx echo.Context) (err error) {
		inParams := make([]reflect.Value, 0)
		if inFlags&handlerHasCtx != 0 {
			inParams = append(inParams, reflect.ValueOf(ctx))
		}
		//has req data
		if inFlags&handlerHasReqData != 0 {
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
			} else {
				// bind
				err = ctx.Bind(req)
			}
			if err != nil {
				return kerrors.ErrBadRequestf("Bad Request:%s ", err)
			}
			if !reqIsPtr {
				req = reflect.ValueOf(req).Elem().Interface()
			}
			//validate
			err = ctx.Validate(req)
			if err != nil {
				return kerrors.ErrBadRequestf("Bad Request:%s ", err)
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
		if !outs[rspErrIdx].IsNil() { // check error
			err = outs[rspErrIdx].Interface().(error)
			if err != nil {
				return err
			}
		}
		var respData interface{}
		if rspDataIdx != -1 {
			rspKind := outs[rspDataIdx].Kind()
			if rspKind == reflect.Ptr || rspKind == reflect.Struct {
				if !(outs[rspDataIdx]).IsNil() {
					respData = outs[rspDataIdx].Interface()
				}
			} else {
				respData = outs[rspDataIdx].Interface()
			}
		}

		if cfg.SkipFormat {
			return ctx.JSON(http.StatusOK, respData)
		}
		var resp interface{}
		switch err.(type) {
		case merrors.Error:
			resp = respData
		default:
			resp = kerrors.OkResult(respData)
		}
		if !ctx.Response().Committed {
			return ctx.JSON(http.StatusOK, resp)
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
