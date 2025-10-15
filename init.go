package web

import (
	"github.com/guestin/kboot"
	"github.com/guestin/kboot/web/internal"
)

func init() {
	kboot.RegisterUnit(ModuleName, _init)
}

func _init(unit kboot.Unit) (kboot.ExecFunc, error) {
	internal.Log = kboot.GetTaggedLogger(ModuleName)
	internal.ZapLog = kboot.GetTaggedZapLogger(ModuleName)
	cfg := new(Config)
	err := kboot.UnmarshalSubConfig(ModuleName, cfg,
		kboot.MustBindEnv(CfgKeyListen),
		kboot.MustBindEnv(CfgKeyDebug),
	)
	if err != nil {
		return nil, err
	}
	return _initEcho(unit, cfg)
}
