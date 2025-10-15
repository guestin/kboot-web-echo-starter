package web

import (
	"github.com/guestin/kboot/web/mid"
)

const (
	ModuleName = "web"

	CfgKeyListen = "listen"
	CfgKeyDebug  = "debug"

	DefaultListenAddress = ":8080"
)

type (
	Config struct {
		ListenAddress string          `toml:"listen" validate:"required" mapstruct:"auth"`
		Debug         bool            `toml:"debug"`
		Auth          *mid.AuthConfig `toml:"auth" validate:"omitnil" mapstruct:"auth"`
		Cors          *mid.CorsConfig `toml:"cors" validate:"omitnil" mapstruct:"cors"`
	}
)
