package web

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type Option interface {
	apply(e *web) error
}

type optionFunc func(w *web) error

func (f optionFunc) apply(w *web) error {
	return f(w)
}

func Use(m ...echo.MiddlewareFunc) {
	_gWeb.options = append(_gWeb.options, optionFunc(func(w *web) error {
		w.echoCtx.Use(m...)
		return nil
	}))
}

type RouteBuilder func(eCtx *echo.Echo)

func WithRouter(fn RouteBuilder) {
	_gWeb.options = append(_gWeb.options, optionFunc(func(w *web) (err error) {
		defer func() {
			pe := recover()
			if pe != nil {
				err = errors.Errorf("panic while register router :%v", pe)
			}
		}()
		fn(w.echoCtx)
		return nil
	}))
}
