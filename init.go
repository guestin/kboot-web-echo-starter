package web

import (
	"github.com/guestin/kboot"
	"github.com/labstack/echo/v4"
)

func init() {
	kboot.RegisterUnit(ModuleName, _init)
	_gWeb = &web{
		ctx:     nil,
		echoCtx: echo.New(),
		cfg:     nil,
		unit:    nil,
		options: make([]Option, 0),
	}
}

func _init(unit kboot.Unit) (kboot.ExecFunc, error) {
	_gWeb.ctx = unit.GetContext()
	_gWeb.unit = unit
	_gWeb.logger = kboot.GetTaggedZapLogger(ModuleName)
	cfg := &Config{
		ListenAddress: DefaultListenAddress,
		Debug:         false,
		Auth:          nil,
		Cors:          nil,
	}
	err := kboot.UnmarshalSubConfig(ModuleName, cfg,
		kboot.MustBindEnv(CfgKeyListen),
		kboot.MustBindEnv(CfgKeyDebug),
	)
	if err != nil {
		return nil, err
	}
	_gWeb.cfg = cfg
	err = _gWeb.Init()
	if err != nil {
		return nil, err
	}
	return func(unit kboot.Unit) kboot.ExitResult {
		<-unit.Done()
		return kboot.ExitResult{
			Code:  0,
			Error: nil,
		}
	}, nil
}
