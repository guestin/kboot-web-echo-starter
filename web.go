package web

import "github.com/labstack/echo/v4"

func EchoCtx() *echo.Echo {
	return _gWeb.echoCtx
}

func GetConfig() *Config {
	return _gWeb.cfg
}

func Start() error {
	return _gWeb.Start()
}

func Shutdown() error {
	return _gWeb.Shutdown()
}
