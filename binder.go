package web

import (
	"bytes"
	"io"

	"github.com/labstack/echo/v4"
)

type _binder struct {
	under *echo.DefaultBinder
}

func (b *_binder) Bind(i interface{}, c echo.Context) error {
	req := c.Request()
	if req.Body != nil && req.ContentLength > 0 {
		reqBody, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		defer func() {
			// reset the body for next bind
			req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}()
		//replayBody, ok := req.Body.(io.ReadSeekCloser)
		//if ok {
		//	restoreFn, err := mio.SaveSeekerPos(replayBody)
		//	if err != nil {
		//		return kerrors.InternalErr(err)
		//	}
		//	defer restoreFn()
		//}
	}
	return b.under.Bind(i, c)
}
