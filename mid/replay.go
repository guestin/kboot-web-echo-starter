package mid

import (
	"bytes"
	"io"

	"github.com/labstack/echo/v4"
)

func ReqBodyReplay() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
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
			return next(c)
		}
	}
}
