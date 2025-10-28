package kerrors

import (
	"fmt"

	"github.com/guestin/mob/merrors"
)

const (
	CodeOk = 0

	CodeOptErr = 2000

	CodeBadRequest    = 4400
	CodeUnauthorized  = 4401
	CodeForbidden     = 4403
	CodeNotFound      = 4404
	CodeDuplicateAdd  = 4409
	CodeInvalidParams = 4422

	CodeInternalServer = 5000

	CodeDbNormalErr       = 6000
	CodeRecordCreateErr   = 6001
	CodeRecordUpdateErr   = 6002
	CodeRecordRetrieveErr = 6003
	CodeRecordDeleteErr   = 6004
	CodeRecordDuplicate   = 6005
)

var codeText = map[int]string{
	CodeOk:            "Success",
	CodeOptErr:        "其他错误",
	CodeUnauthorized:  "用户未登录或登录已失效",
	CodeForbidden:     "无权限",
	CodeNotFound:      "无记录",
	CodeDuplicateAdd:  "重复添加",
	CodeBadRequest:    "请求参数不正确",
	CodeInvalidParams: "请求参数不正确",

	CodeInternalServer: "服务异常",

	CodeDbNormalErr:       "数据库操作失败",
	CodeRecordCreateErr:   "数据添加失败",
	CodeRecordUpdateErr:   "数据更新失败",
	CodeRecordRetrieveErr: "数据查询失败",
	CodeRecordDeleteErr:   "数据删除失败",
	CodeRecordDuplicate:   "记录重复",
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

func WrapSensitiveErr(err merrors.Error) merrors.Error {
	if err == nil {
		return nil
	}
	if err.GetCode() < CodeInternalServer {
		return err
	}
	return merrors.Errorf0(err.GetCode(), fmt.Sprintf("%s,请联系系统管理员处理", CodeText(err.GetCode())))
}
