# Config

```toml
[web]
# only support sqlite or postgres
listen = "0.0.0.0:8080"
debug = true
```

## Usage

```go
package apis

import (
	"github.com/guestin/kboot/web"
	"github.com/guestin/kboot/web/mid"
)

func init() {
	web.Router(routeBuilder)
	web.DependsOn("db")
}

func routeBuilder(eCtx *echo.Echo) error {
	eCtx.GET("/echo", mid.Wrap(Echo))
	return nil
}

```
