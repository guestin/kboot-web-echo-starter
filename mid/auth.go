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
	AuthSessionInfo interface {
		UserId() string
		ExpireAt() int64
	}
	AuthContext interface {
		IsAnonymous() bool
		GetUserId() string
		GetSessionId() string
		ExpireAt() int64
		ClientIp() string
		ClientUA() string
		SessionInfo() AuthSessionInfo
	}
	SessionLoadFunc func(ctx echo.Context, sessionId string) (AuthSessionInfo, error)
	AuthConfig      struct {
		Enabled              bool     `toml:"enabled" json:"enabled" mapstructure:"enabled"` //是否启用，启用后将解析session info
		Whitelist            []string `toml:"whitelist" json:"whitelist" mapstructure:"whitelist"`
		SessionIdKey         string   `toml:"sessionIdKey" json:"sessionIdKey" mapstructure:"sessionIdKey"`
		SessionExpireInHours int      `toml:"sessionExpireInHours" json:"sessionExpireInHours" validate:"gte=0,lte=720" mapstructure:"sessionExpireInHours"`
		SessionLoadFunc      SessionLoadFunc
	}
)

var DefaultAuthConfig = AuthConfig{
	Enabled:      false,
	Whitelist:    []string{},
	SessionIdKey: "kt-session-id",
}

type _authCtx struct {
	isAnonymous bool
	userId      string
	sessionId   string
	expireAt    int64
	clientIp    string
	clientUA    string
	sessionInfo AuthSessionInfo
}

type _anonymousSession struct {
	userId string
}

func (this *_anonymousSession) UserId() string {
	return ""
}

func (this *_anonymousSession) ExpireAt() int64 {
	return 0
}

func (this *_authCtx) IsAnonymous() bool {
	return this.isAnonymous
}

func (this *_authCtx) GetUserId() string {
	return this.userId
}

func (this *_authCtx) GetSessionId() string {
	return this.sessionId
}

func (this *_authCtx) ExpireAt() int64 {
	return this.expireAt
}

func (this *_authCtx) ClientIp() string {
	return this.clientIp
}

func (this *_authCtx) ClientUA() string {
	return this.clientUA
}

func (this *_authCtx) SessionInfo() AuthSessionInfo {
	return this.sessionInfo
}

func (this *_authCtx) reset(realIp string, ua string) {
	now := time.Now()
	randomId := fmt.Sprintf("ANONYMOUS_%s", now.Format("060102150405.000000"))
	this.isAnonymous = true
	this.sessionId = ""
	this.userId = randomId
	this.sessionInfo = &_anonymousSession{userId: randomId}
	this.expireAt = time.Now().Add(time.Hour * 24).Unix()
	this.clientIp = realIp
	this.clientUA = ua
}

func CurrentAuthContext(ctx echo.Context) AuthContext {
	return ctx.Get(CtxCallerInfoKey).(AuthContext)
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
	authCtxPool := &sync.Pool{
		New: func() interface{} { return new(_authCtx) },
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			authCtx := authCtxPool.Get().(*_authCtx)
			authCtx.reset(ctx.RealIP(), ctx.Request().UserAgent())
			defer func() {
				authCtxPool.Put(authCtx)
				ctx.Set(CtxCallerInfoKey, nil)
			}()
			ctx.Set(CtxCallerInfoKey, authCtx)
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
			sessionId := ctx.Request().Header.Get(config.SessionIdKey)
			if len(sessionId) == 0 {
				sessionId = ctx.QueryParam(config.SessionIdKey)
			}
			if len(sessionId) == 0 && !ignore {
				return kerrors.ErrUnauthorized()
			}
			authCtx.sessionId = sessionId
			if config.SessionLoadFunc != nil && len(sessionId) > 0 {
				sessionInfo, err := config.SessionLoadFunc(ctx, sessionId)
				if err == nil {
					authCtx.isAnonymous = false
					authCtx.userId = sessionInfo.UserId()
					authCtx.expireAt = sessionInfo.ExpireAt()
					authCtx.sessionInfo = sessionInfo
				} else {
					if !ignore {
						return kerrors.ErrUnauthorized()
					}
				}
			}
			return next(ctx)
		}
	}
}
