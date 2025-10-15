package mid

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/guestin/kboot-web-echo-starter/kerrors"
	"github.com/guestin/mob"
	"github.com/labstack/echo/v4"
)

type (
	IUserData interface {
		GetUid() string
		GetExpireAt() int64
	}

	AuthSession struct {
		IsAnonymous bool
		Uid         string
		SessionId   string
		ExpireAt    int64
		ClientIp    string
		ClientUA    string
		UserData    IUserData
	}
	SessionCheckCallBack func(ctx echo.Context, sessionId string) (IUserData, error)
	AuthConfig           struct {
		Enabled              bool     `toml:"enabled" mapstructure:"enabled"` //是否启用，启用后将解析session info
		Whitelist            []string `toml:"whitelist" mapstructure:"whitelist"`
		SessionIdKey         string   `toml:"sessionIdKey" mapstructure:"sessionIdKey"`
		SessionExpireInHours int      `toml:"sessionExpireInHours" validate:"gte=0,lte=720" mapstructure:"sessionExpireInHours"`
	}
)

var DefaultAuthConfig = AuthConfig{
	Enabled:      false,
	Whitelist:    []string{},
	SessionIdKey: "kt-session-id",
}
var _AuthCheckCbFn SessionCheckCallBack = nil

func SetupAuthCheckFn(cb SessionCheckCallBack) {
	_AuthCheckCbFn = cb
}

func (this *AuthSession) reset(realIp string, ua string) {
	now := time.Now()
	randomId := fmt.Sprintf("ANONYMOUS_%s_%s", now.Format("060102150405.000000"), realIp)
	this.IsAnonymous = true
	this.SessionId = ""
	this.Uid = randomId
	this.UserData = nil
	this.ClientIp = realIp
	this.ClientUA = ua
}

func CurrentSession(ctx echo.Context) *AuthSession {
	return ctx.Get(CtxCallerInfoKey).(*AuthSession)
}

func CurrentSessionData[T any](ctx echo.Context) T {
	return ctx.Get(CtxCallerInfoKey).(*AuthSession).UserData.(T)
}

func Auth() echo.MiddlewareFunc {
	return AuthWithConfig(DefaultAuthConfig)
}

func AuthWithConfig(config AuthConfig) echo.MiddlewareFunc {
	if strings.Trim(config.SessionIdKey, " ") == "" {
		config.SessionIdKey = DefaultAuthConfig.SessionIdKey
	}
	excludePathSet := mob.NewConcurrentSet()
	excludeRegList := make([]*regexp.Regexp, 0)
	for i := range config.Whitelist {
		p := config.Whitelist[i]
		if excludePathSet.Add(p) {
			reg, err := regexp.Compile(p)
			if err != nil {
				panic(fmt.Sprintf("witelist path %s is not a valid reg path", p))
			}
			excludeRegList = append(excludeRegList, reg)
		}
	}
	sessionPool := &sync.Pool{
		New: func() interface{} { return new(AuthSession) },
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			sessionInfo := sessionPool.Get().(*AuthSession)
			sessionInfo.reset(ctx.RealIP(), ctx.Request().UserAgent())
			defer func() {
				sessionPool.Put(sessionInfo)
				ctx.Set(CtxCallerInfoKey, nil)
			}()
			reqPath := ctx.Request().URL.String()
			ignore := false
			if !config.Enabled {
				ignore = true
			} else {
				for i := range excludeRegList {
					if excludeRegList[i].MatchString(reqPath) {
						ignore = true
						break
					}
				}
			}
			token := ctx.Request().Header.Get(config.SessionIdKey)
			if len(token) == 0 {
				token = ctx.QueryParam(config.SessionIdKey)
			}
			if len(token) == 0 && !ignore {
				return kerrors.ErrUnauthorized()
			}
			sessionInfo.SessionId = token
			var userData IUserData
			var err error
			if len(token) > 0 {
				if _AuthCheckCbFn != nil {
					userData, err = _AuthCheckCbFn(ctx, token)
					if err != nil && !ignore {
						return err
					}
				}
			}
			if userData != nil {
				sessionInfo.IsAnonymous = false
				sessionInfo.Uid = userData.GetUid()
				sessionInfo.UserData = userData
				sessionInfo.ExpireAt = userData.GetExpireAt()
			}
			ctx.Set(CtxCallerInfoKey, sessionInfo)
			return next(ctx)
		}
	}
}
