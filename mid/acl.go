package mid

import (
	"sync"

	"github.com/guestin/kboot"
	"github.com/guestin/kboot-web-echo-starter/kerrors"
	"github.com/labstack/echo/v4"
)

type (
	ACLContext interface {
		AllPermissions() []ACLPermission
		MatchedPermissions() []ACLPermission
	}
	ACLPermission interface {
		Match(ctx echo.Context) bool
	}
	ACLPermissionLoadFunc func(ctx echo.Context) ([]ACLPermission, error)
	ACLConfig             struct {
		Enabled               bool `toml:"enabled" mapstructure:"enabled"`
		Skipper               Skipper
		BeforeFunc            BeforeFunc
		ACLPermissionLoadFunc ACLPermissionLoadFunc
	}
)

var DefaultACLConfig = ACLConfig{
	Enabled: false,
}

type _aclCtx struct {
	allPermissions    []ACLPermission
	matchedPermission []ACLPermission
}

func (this *_aclCtx) AllPermissions() []ACLPermission {
	return this.allPermissions
}

func (this *_aclCtx) MatchedPermissions() []ACLPermission {
	return this.matchedPermission
}

func (this *_aclCtx) reset() {
	this.allPermissions = make([]ACLPermission, 0)
	this.matchedPermission = make([]ACLPermission, 0)
}

func CurrentACLContext(ctx echo.Context) ACLContext {
	return ctx.Get(CtxAclKey).(ACLContext)
}

func ACL(config ACLConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.ACLPermissionLoadFunc == nil {
		kboot.GetTaggedZapLogger("mid.acl").Panic("ACL permission loader not set")
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		ctxPool := &sync.Pool{
			New: func() interface{} { return new(_aclCtx) },
		}
		return func(ctx echo.Context) error {
			aclCtx := ctxPool.Get().(*_aclCtx)
			aclCtx.reset()
			defer func() {
				ctxPool.Put(aclCtx)
				ctx.Set(CtxAclKey, nil)
			}()
			ctx.Set(CtxAclKey, aclCtx)
			if !config.Enabled {
				return next(ctx)
			}
			if config.Skipper != nil && config.Skipper(ctx) {
				return next(ctx)
			}
			if config.BeforeFunc != nil {
				err := config.BeforeFunc(ctx)
				if err != nil {
					return err
				}
			}
			if config.ACLPermissionLoadFunc != nil {
				permissions, err := config.ACLPermissionLoadFunc(ctx)
				if err != nil {
					return err
				}
				aclCtx.allPermissions = permissions[:]
				for i := range permissions {
					perm := permissions[i]
					if perm.Match(ctx) {
						aclCtx.matchedPermission = append(aclCtx.matchedPermission, perm)
					}
				}
				if len(aclCtx.matchedPermission) == 0 {
					return kerrors.ErrForbidden()
				}
			}
			return next(ctx)
		}
	}
}
