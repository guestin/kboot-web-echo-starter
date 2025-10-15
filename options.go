package web

import (
	"github.com/labstack/echo/v4"
)

var _options = make([]Option, 0)

type Option interface {
	apply(e *echo.Echo) error
}

type optionFunc func(e *echo.Echo) error

func (f optionFunc) apply(e *echo.Echo) error {
	return f(e)
}

type RouteBuilder func(eCtx *echo.Echo) error

func Use(m ...echo.MiddlewareFunc) {
	_options = append(_options, optionFunc(func(e *echo.Echo) error {
		e.Use(m...)
		return nil
	}))
}

func Router(fn RouteBuilder) {
	_options = append(_options, optionFunc(func(e *echo.Echo) error {
		return fn(e)
	}))
}
