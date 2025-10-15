package mid

type CorsConfig struct {
	Enabled          bool     `toml:"enabled" mapstructure:"enabled"`
	AllowOrigins     []string `toml:"allowOrigins" validate:"omitempty,dive,required" mapstruct:"allowOrigins"`
	AllowMethods     []string `toml:"allowMethods" validate:"omitempty,dive,required" mapstruct:"allowMethods"`
	AllowHeaders     []string `toml:"allowHeaders" validate:"omitempty,dive,required" mapstruct:"allowHeaders"`
	ExposeHeaders    []string `toml:"exposeHeaders" validate:"omitempty,dive,required" mapstruct:"exposeHeaders"`
	AllowCredentials bool     `toml:"allowCredentials" mapstructure:"allowCredentials"`
	MaxAge           int      `toml:"maxAge" validate:"gte=0,lte=86400" mapstructure:"maxAge"`
}
