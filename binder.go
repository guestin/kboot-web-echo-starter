package web

import (
	"io"

	"github.com/labstack/echo/v4"
)

type _binder struct {
	under *echo.DefaultBinder
}

func (b *_binder) Bind(i interface{}, c echo.Context) error {
	req := c.Request()
	if req.Body != nil && req.ContentLength > 0 {
		replayBody, ok := req.Body.(io.ReadSeekCloser)
		if ok {
			defer func() {
				_, _ = replayBody.Seek(0, io.SeekStart)
			}()
		}
	}
	return b.under.Bind(i, c)
}
