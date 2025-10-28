package mid

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type (
	AuditContext interface {
		OverrideUserId(userId string)
		Set(key string, value interface{}) AuditContext
		Get(key string) interface{}
		SetResourceId(resourceId string)
		GetResourceId() string
		WithError(err error)
		UserId() string
		ClientIp() string
		ClientUA() string
		Begin() time.Time
		UserDataJson() string
		DumpError() string
	}
	FlushFunc   func(ctx echo.Context)
	AuditConfig struct {
		Enabled   bool `toml:"enabled" mapstructure:"enabled"`
		Skipper   Skipper
		FlushFunc FlushFunc
	}
)

type _auditCtx struct {
	userData   map[string]interface{}
	resourceId string
	errs       []error
	begin      time.Time
	userId     string
	clientIp   string
	clientUA   string
}

func (this *_auditCtx) OverrideUserId(userId string) {
	this.userId = userId
}

func (this *_auditCtx) SetResourceId(resourceId string) {
	this.resourceId = resourceId
}

func (this *_auditCtx) GetResourceId() string {
	return this.resourceId
}

func (this *_auditCtx) Set(key string, value interface{}) AuditContext {
	if value != nil {
		this.userData[key] = value
	}
	return this
}

func (this *_auditCtx) Get(key string) interface{} {
	return this.userData[key]
}

func (this *_auditCtx) WithError(err error) {
	this.errs = append(this.errs, err)
}

func (this *_auditCtx) UserId() string {
	return this.userId
}

func (this *_auditCtx) ClientIp() string {
	return this.clientIp
}

func (this *_auditCtx) ClientUA() string {
	return this.clientUA
}

func (this *_auditCtx) Begin() time.Time {
	return this.begin
}
func (this *_auditCtx) UserDataJson() string {
	detailStr, _ := json.Marshal(this.userData)
	return string(detailStr)
}

func (this *_auditCtx) DumpError() string {
	errSb := strings.Builder{}
	for _, err := range this.errs {
		errSb.WriteString(err.Error())
	}
	return errSb.String()
}

func (this *_auditCtx) reset() {
	this.userData = make(map[string]interface{})
	this.errs = make([]error, 0)
	this.begin = time.Now()
	this.userId = ""
	this.clientIp = ""
	this.clientUA = ""
}

func CurrentAuditContext(ctx echo.Context) AuditContext {
	return ctx.Get(CtxAuditKey).(AuditContext)
}

func Audit(config AuditConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	ctxPool := &sync.Pool{
		New: func() interface{} { return new(_auditCtx) },
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			auditCtx := ctxPool.Get().(*_auditCtx)
			auditCtx.reset()
			auditCtx.userId = CurrentAuthContext(ctx).GetUserId()
			auditCtx.clientUA = ctx.Request().UserAgent()
			auditCtx.clientIp = ctx.RealIP()
			defer func() {
				ctxPool.Put(auditCtx)
				ctx.Set(CtxAuditKey, nil)
			}()
			ctx.Set(CtxAuditKey, auditCtx)
			if !config.Enabled {
				return next(ctx)
			}
			if config.Skipper(ctx) {
				return next(ctx)
			}
			err := next(ctx)
			if err != nil {
				auditCtx.WithError(err)
			}
			if config.FlushFunc != nil {
				config.FlushFunc(ctx)
			}
			return err
		}
	}
}
