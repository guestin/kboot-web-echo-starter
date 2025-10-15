package web

import (
	"github.com/guestin/kboot-web-echo-starter/mid"
)

const (
	ModuleName = "web"

	CfgKeyListen = "listen"
	CfgKeyDebug  = "debug"

	DefaultListenAddress = ":20808"
)

type (
	Config struct {
		ListenAddress string          `toml:"listen" validate:"required" mapstruct:"auth"`
		Debug         bool            `toml:"debug"`
		Auth          *mid.AuthConfig `toml:"auth" validate:"omitnil" mapstruct:"auth"`
		Cors          *mid.CorsConfig `toml:"cors" validate:"omitnil" mapstruct:"cors"`
	}
)
