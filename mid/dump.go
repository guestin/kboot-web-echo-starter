package mid

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mattn/go-isatty"
)

type DumpOption uint64

// noinspection ALL
const (
	LogNone DumpOption = 1 << iota
	LogReqHeader
	LogRespHeader
	LogForm
	LogMultipartForm
	LogJson
	LogHtml
	LogTextPlain
	LogXml
	LogJs
	LogAll
)

// noinspection ALL
const (
	charsetUTF8         = "charset=UTF-8"
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
)

// MIME types
// noinspection ALL
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMETextXML                          = "text/xml"
	MIMETextXMLCharsetUTF8               = MIMETextXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

// noinspection ALL
func Dump(option DumpOption) echo.MiddlewareFunc {
	return DumpWithConfig(LoggerConfig{Option: option})
}

func DumpWithConfig(conf LoggerConfig) echo.MiddlewareFunc {
	option := conf.Option
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultLogFormatter
	}
	out := conf.Output
	if out == nil {
		out = DefaultWriter
	}

	isTerm := true
	if w, ok := out.(*os.File); !ok || os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd())) {
		isTerm = false
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			requestId := ctx.Response().Header().Get(echo.HeaderXRequestID)
			start := time.Now()
			path := ctx.Request().URL.Path
			raw := ctx.Request().URL.RawQuery
			param := LogFormatterParams{
				Request:   ctx.Request(),
				isTerm:    isTerm,
				RequestId: requestId,
			}
			if option&LogReqHeader != 0 {
				param.RequestHeader = ctx.Request().Header
			}
			reqCt := ctx.Request().Header.Get(echo.HeaderContentType)
			if ctx.Request().ContentLength > 0 && shouldDump(option, reqCt) {
				reqBody := make([]byte, 0)
				reqBody, _ = io.ReadAll(ctx.Request().Body)
				ctx.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody)) // reset
				param.RequestBody = reqBody
			}
			resBody := new(bytes.Buffer)
			if shouldCacheBody(option) {
				mw := io.MultiWriter(ctx.Response().Writer, resBody)
				writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: ctx.Response().Writer}
				ctx.Response().Writer = writer
			}
			err := next(ctx)
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)
			param.ClientIP = ctx.RealIP()
			param.Method = ctx.Request().Method
			param.StatusCode = ctx.Response().Status
			param.BodySize = ctx.Response().Size
			if raw != "" {
				path = path + "?" + raw
			}
			param.Path = path
			if option&LogRespHeader != 0 {
				param.ResponseHeader = ctx.Response().Header()
			}
			resCt := ctx.Response().Header().Get(HeaderContentType)
			if shouldDump(option, resCt) {
				param.ResponseBody = resBody.Bytes()
			}
			_, _ = fmt.Fprint(out, formatter(param))
			return err
		}
	}
}

func shouldDump(option DumpOption, contentType string) bool {
	return (option&LogAll != 0 ||
		option&LogForm != 0 && strings.Contains(contentType, MIMEApplicationForm)) ||
		(option&LogMultipartForm != 0 && strings.Contains(contentType, MIMEMultipartForm)) ||
		(option&LogJson != 0 && strings.Contains(contentType, MIMEApplicationJSON)) ||
		(option&LogHtml != 0 && strings.Contains(contentType, MIMETextHTML)) ||
		(option&LogTextPlain != 0 && strings.Contains(contentType, MIMETextPlain)) ||
		(option&LogXml != 0 && strings.Contains(contentType, MIMETextXML)) ||
		(option&LogJs != 0 && strings.Contains(contentType, MIMEApplicationJavaScript))
}

func shouldCacheBody(option DumpOption) bool {
	return option&LogAll != 0 ||
		option&LogForm != 0 ||
		option&LogMultipartForm != 0 ||
		option&LogJson != 0 ||
		option&LogHtml != 0 ||
		option&LogTextPlain != 0 ||
		option&LogXml != 0 ||
		option&LogJs != 0
}

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

var DefaultWriter io.Writer = os.Stdout

// LoggerConfig defines the config for Dump middleware.
type LoggerConfig struct {
	// Optional. Default value is gin.defaultLogFormatter
	Formatter LogFormatter

	// Output is a writer where logs are written.
	// Optional. Default value is gin.DefaultWriter.
	Output io.Writer

	Option DumpOption
}

type consoleColorModeValue int

const (
	autoColor consoleColorModeValue = iota
	disableColor
	forceColor
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

var consoleColorMode = autoColor

type LogFormatter func(params LogFormatterParams) string

type LogFormatterParams struct {
	Request        *http.Request
	TimeStamp      time.Time
	StatusCode     int
	Latency        time.Duration
	ClientIP       string
	Method         string
	Path           string
	isTerm         bool
	BodySize       int64
	RequestId      string
	RequestBody    []byte
	RequestHeader  http.Header
	ResponseBody   []byte
	ResponseHeader http.Header
}

func (p *LogFormatterParams) FormattedRequestStr() string {
	return formatHeadersAndBody("<<<<<<<<<<\n", p.RequestHeader, p.RequestBody)
}

func (p *LogFormatterParams) FormattedResponseStr() string {
	return formatHeadersAndBody(">>>>>>>>>>\n", p.ResponseHeader, p.ResponseBody)
}
func formatHeadersAndBody(prefix string, header http.Header, body []byte) string {
	buf := bytes.Buffer{}
	if len(header) > 0 || len(body) > 0 {
		buf.WriteString(prefix)
	}
	if len(header) > 0 {
		buf.WriteString("Headers:\n")
		for k, v := range header {
			if strings.Compare(strings.ToLower(k), strings.ToLower(HeaderAuthorization)) == 0 {
				buf.WriteString(fmt.Sprintf("  %s : ****** \n", k))
			} else {
				buf.WriteString(fmt.Sprintf("  %s : %s \n", k, strings.Join(v, " ")))
			}
		}
	}
	if len(body) > 0 {
		buf.WriteString(fmt.Sprintf("Body:\n%s", string(body)))
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		buf.WriteString("\n")
	}
	return buf.String()
}

func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

func (p *LogFormatterParams) MethodColor() string {
	method := p.Method
	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

func (p *LogFormatterParams) ResetColor() string {
	return reset
}

// IsOutputColor indicates whether can colors be outputted to the log.
func (p *LogFormatterParams) IsOutputColor() bool {
	return consoleColorMode == forceColor || (consoleColorMode == autoColor && p.isTerm)
}

var defaultLogFormatter = func(param LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency - param.Latency%time.Second
	}
	return fmt.Sprintf("[%s] %v |%s %3d %s| %13v | %15s |%s %-7s %s %s\n"+
		"%s"+
		"%s",
		param.RequestId,
		param.TimeStamp.Format("2006-01-02 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.FormattedRequestStr(),
		param.FormattedResponseStr(),
	)
}

//goland:noinspection ALL
func DisableConsoleColor() {
	consoleColorMode = disableColor
}

//goland:noinspection ALL
func ForceConsoleColor() {
	consoleColorMode = forceColor
}
