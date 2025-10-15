package kerrors

const (
	CodeOk            = 0
	CodeBadRequest    = 4400
	CodeUnauthorized  = 4401
	CodeForbidden     = 4403
	CodeNotFound      = 4404
	CodeDuplicateAdd  = 4409
	CodeInvalidParams = 4422
	CodeOptErr        = 2000

	CodeInternalServer = 5000

	CodeInternalErr       = 6000
	CodeRecordCreateErr   = 6001
	CodeRecordUpdateErr   = 6002
	CodeRecordRetrieveErr = 6003
	CodeRecordDeleteErr   = 6004
)

var codeText = map[int]string{
	CodeOk:             "Success",
	CodeUnauthorized:   "用户未登录或登录已失效",
	CodeForbidden:      "无权限",
	CodeNotFound:       "无记录",
	CodeDuplicateAdd:   "重复添加",
	CodeBadRequest:     "请求参数不正确",
	CodeInvalidParams:  "请求参数不正确",
	CodeInternalServer: "服务异常",
	CodeOptErr:         "其他错误",

	CodeInternalErr:       "服务器内部错误",
	CodeRecordCreateErr:   "添加失败",
	CodeRecordUpdateErr:   "更新失败",
	CodeRecordRetrieveErr: "查询失败",
	CodeRecordDeleteErr:   "删除失败",
}

func CodeText(code int) string {
	if v, ok := codeText[code]; ok {
		return v
	}
	return "未知错误"
}

func HttpStatus2Code(status int) int {
	return (status/100)*1000 + status
}
