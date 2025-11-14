package mid

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/guestin/log"
	"github.com/guestin/mob"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	LogBodyOption uint64

	LoggerConfig struct {
		Skipper           Skipper
		Logger            log.ZapLog
		LogReqHeader      *bool
		HideHeader        []string
		LogReqBody        *bool
		LogReqBodyOption  LogBodyOption
		LogRespHeader     *bool
		LogRespBody       *bool
		LogRespBodyOption LogBodyOption
	}
)

const (
	LogBodyForm LogBodyOption = 1 << iota
	LogBodyMultipartForm
	LogBodyJson
	LogBodyHtml
	LogBodyTextPlain
	LogBodyXml
	LogBodyJs
	LogBodyAll
)

func teaBool(v bool) *bool {
	return &v
}

var (
	DefaultLoggerConfig = LoggerConfig{
		Skipper:           DefaultSkipper,
		Logger:            nil,
		LogReqHeader:      teaBool(true),
		HideHeader:        make([]string, 0),
		LogReqBody:        teaBool(true),
		LogReqBodyOption:  LogBodyForm | LogBodyXml | LogBodyJson,
		LogRespHeader:     teaBool(true),
		LogRespBody:       teaBool(true),
		LogRespBodyOption: LogBodyForm | LogBodyXml | LogBodyJson,
	}
)

func Logger(logger log.ZapLog) echo.MiddlewareFunc {
	return LoggerWithConfig(LoggerConfig{
		Logger: logger,
	})
}

func LoggerWithConfig(config LoggerConfig) echo.MiddlewareFunc {
	if config.Logger == nil {
		panic("Logger must not be nil")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.LogReqBody == nil {
		config.LogReqBody = DefaultLoggerConfig.LogReqBody
	}
	if config.LogReqBodyOption == 0 {
		config.LogReqBodyOption = DefaultLoggerConfig.LogReqBodyOption
	}
	if config.LogReqHeader == nil {
		config.LogReqHeader = DefaultLoggerConfig.LogReqHeader
	}
	if config.LogRespBody == nil {
		config.LogRespBody = DefaultLoggerConfig.LogRespBody
	}
	if config.LogRespBodyOption == 0 {
		config.LogRespBodyOption = DefaultLoggerConfig.LogRespBodyOption
	}
	if config.LogRespHeader == nil {
		config.LogRespHeader = DefaultLoggerConfig.LogRespHeader
	}
	hideHeader := mob.NewSet()
	for _, h := range config.HideHeader {
		hideHeader.Add(strings.ToLower(h))
	}
	config.Logger = config.Logger.With(log.WithZapOptions(zap.WithCaller(false)))
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}
			begin := time.Now()
			traceId := GetTraceId(ctx)
			logger := config.Logger
			if traceId != "" {
				logger = config.Logger.With(
					log.UseSubTag(log.NewFixStyleText(traceId, log.Blue, false)))
			}
			relUrl := ctx.Request().URL.Path
			method := ctx.Request().Method
			rawQuery := ctx.Request().URL.RawQuery
			if rawQuery != "" {
				relUrl = relUrl + "?" + rawQuery
			}
			clientIp := ctx.RealIP()
			logger.Info(fmt.Sprintf("<<<<<<<<<< %s | %s %s", clientIp, method, relUrl))
			if *config.LogReqHeader {
				loggerHeader(ctx.Request().Header, logger, hideHeader)
			}
			reqCt := ctx.Request().Header.Get(echo.HeaderContentType)
			if *config.LogReqBody && ctx.Request().ContentLength > 0 && shouldLogBody(config.LogReqBodyOption, reqCt) {
				reqBody := make([]byte, 0)
				reqBody, _ = io.ReadAll(ctx.Request().Body)
				ctx.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody)) // reset
				if strings.Contains(reqCt, MIMEApplicationJSON) {
					buf := new(bytes.Buffer)
					if ce := json.Compact(buf, reqBody); ce == nil {
						reqBody = buf.Bytes()
					}
				}
				bodyLines := strings.Split(string(reqBody), "\n")
				logger.Debug("Body:")
				for _, line := range bodyLines {
					logger.Debug(line)
				}
			}
			resBody := new(bytes.Buffer)
			if *config.LogRespBody {
				mw := io.MultiWriter(ctx.Response().Writer, resBody)
				writer := &loggerBodyHijackWriter{Writer: mw, ResponseWriter: ctx.Response().Writer}
				ctx.Response().Writer = writer
			}
			err := next(ctx)
			if err != nil {
				ctx.Error(err)
			}
			latency := time.Now().Sub(begin)
			statusCode := ctx.Response().Status
			bodySize := ctx.Response().Size
			output := logger.Info
			if statusCode >= http.StatusOK && statusCode <= http.StatusMultipleChoices {
			} else {
				output = logger.Warn
			}
			output(fmt.Sprintf(">>>>>>>>>> %s | %3d | %s | %s %s",
				clientIp,
				statusCode,
				latency.String(),
				method,
				relUrl),
			)
			if *config.LogRespHeader {
				loggerHeader(ctx.Response().Header(), logger, hideHeader)
			}
			resCt := ctx.Response().Header().Get(echo.HeaderContentType)
			if *config.LogRespBody && bodySize > 0 && shouldLogBody(config.LogRespBodyOption, resCt) {
				bodyLines := strings.Split(resBody.String(), "\n")
				logger.Debug("Body:")
				for _, line := range bodyLines {
					logger.Debug(line)
				}
			}
			return nil
		}
	}
}

func shouldLogBody(option LogBodyOption, contentType string) bool {
	return (option&LogBodyAll != 0 ||
		option&LogBodyForm != 0 && strings.Contains(contentType, MIMEApplicationForm)) ||
		(option&LogBodyMultipartForm != 0 && strings.Contains(contentType, MIMEMultipartForm)) ||
		(option&LogBodyJson != 0 && strings.Contains(contentType, MIMEApplicationJSON)) ||
		(option&LogBodyHtml != 0 && strings.Contains(contentType, MIMETextHTML)) ||
		(option&LogBodyTextPlain != 0 && strings.Contains(contentType, MIMETextPlain)) ||
		(option&LogBodyXml != 0 && strings.Contains(contentType, MIMETextXML)) ||
		(option&LogBodyJs != 0 && strings.Contains(contentType, MIMEApplicationJavaScript))
}

func hideString(v string) string {
	l := len(v)
	switch l {
	case 1:
		return "*"
	case 2:
		return fmt.Sprintf("%s*", v[:1])
	case 3, 4, 5, 6:
		return fmt.Sprintf("%s%s%s", v[:1], strings.Repeat("*", len(v)-2), v[len(v)-1:])
	case 7, 8, 9, 10:
		return fmt.Sprintf("%s%s%s", v[:3], strings.Repeat("*", len(v)-6), v[len(v)-3:])
	default:
		return fmt.Sprintf("%s%s%s", v[:4], strings.Repeat("*", len(v)-8), v[len(v)-4:])
	}
}

func loggerHeader(header http.Header, logger log.ZapLog, hide mapset.Set) {
	if len(header) > 0 {
		logger.Debug("Headers:")
		for k, v := range header {
			if hide.Contains(strings.ToLower(k)) {
				vv := strings.Join(v, " ")
				if len(vv) > 8 {
					logger.Debug(fmt.Sprintf("  %s : %s", k, hideString(vv)))
				}
			} else {
				logger.Debug(fmt.Sprintf("  %s : %s ", k, strings.Join(v, " ")))
			}
		}
	}
}

type loggerBodyHijackWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *loggerBodyHijackWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggerBodyHijackWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *loggerBodyHijackWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *loggerBodyHijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
