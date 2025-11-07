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
		ListenAddress string              `toml:"listen" validate:"required" mapstruct:"listen"`
		Debug         bool                `toml:"debug" mapstructure:"debug"`
		Auth          mid.AuthConfig      `toml:"auth" validate:"omitempty" mapstruct:"auth"`
		ACL           mid.ACLConfig       `toml:"acl" validate:"omitempty" mapstruct:"acl"`
		Audit         mid.AuditConfig     `toml:"audit" validate:"omitempty" mapstruct:"audit"`
		Trace         mid.RequestIDConfig `toml:"trace" validate:"omitempty" mapstruct:"trace"`
	}
)
