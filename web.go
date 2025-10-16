package web

func Start() error {
	return _gWeb.Start()
}

func Shutdown() error {
	return _gWeb.Shutdown()
}
