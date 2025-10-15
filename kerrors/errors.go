package kerrors

import (
	"fmt"

	"github.com/guestin/mob/merrors"
)

func OkResult(data interface{}) merrors.Error {
	return merrors.NewError().SetCode(CodeOk).SetMsg(CodeText(CodeOk)).SetData(data)
}

func NewErr(code int, msg ...interface{}) merrors.Error {
	if len(msg) > 0 {
		if res, ok := msg[0].(merrors.Error); ok {
			return res
		} else {
			return merrors.NewError().SetCode(code).SetMsg(fmt.Sprint(msg...))
		}
	}
	return merrors.NewError().SetCode(code).SetMsg(CodeText(code))
}

func Errorf(code int, format string, arg ...interface{}) merrors.Error {
	return merrors.NewError().SetCode(code).SetMsg(fmt.Sprintf(format, arg...))
}

//goland:noinspection ALL
func ErrOK(msg ...interface{}) merrors.Error {
	return NewErr(CodeOk, msg...)
}

//goland:noinspection ALL
func ErrOpt(msg ...interface{}) merrors.Error {
	return NewErr(CodeOptErr, msg...)
}

//goland:noinspection ALL
func ErrOptf(format string, arg ...interface{}) merrors.Error {
	return Errorf(CodeOptErr, format, arg...)
}

//goland:noinspection ALL
func ErrExist(msg ...interface{}) merrors.Error {
	return NewErr(CodeDuplicateAdd, msg...)
}

//goland:noinspection ALL
func ErrExistf(format string, arg ...interface{}) merrors.Error {
	return Errorf(CodeDuplicateAdd, format, arg...)
}

//goland:noinspection ALL
func ErrNotExist(msg ...interface{}) merrors.Error {
	return NewErr(CodeNotFound, msg...)
}

//goland:noinspection ALL
func ErrNotExistf(format string, arg ...interface{}) merrors.Error {
	return Errorf(CodeNotFound, format, arg...)
}

//region  basic http error

// noinspection ALL
func ErrBadRequest(msg ...interface{}) merrors.Error {
	return NewErr(CodeBadRequest, msg...)
}

//goland:noinspection ALL
func ErrBadRequestf(format string, arg ...interface{}) merrors.Error {
	return Errorf(CodeBadRequest, format, arg...)
}

//goland:noinspection ALL
func ErrUnauthorized(msg ...interface{}) merrors.Error {
	return NewErr(CodeUnauthorized, msg...)
}

//goland:noinspection ALL
func ErrUnauthorizedf(format string, arg ...interface{}) merrors.Error {
	return Errorf(CodeUnauthorized, format, arg...)
}

//goland:noinspection ALL
func ErrForbidden(msg ...interface{}) merrors.Error {
	return NewErr(CodeForbidden, msg...)
}

//goland:noinspection ALL
func ErrForbiddenf(format string, arg ...interface{}) merrors.Error {
	return Errorf(CodeForbidden, format, arg...)
}

//goland:noinspection ALL
func ErrDuplicateAdd(msg ...interface{}) merrors.Error {
	return NewErr(CodeDuplicateAdd, msg...)
}

//goland:noinspection ALL
func ErrDuplicateAddf(format string, arg ...interface{}) merrors.Error {
	return Errorf(CodeDuplicateAdd, format, arg...)
}

//goland:noinspection ALL
func ErrInvalidParams(msg ...interface{}) merrors.Error {
	return NewErr(CodeInvalidParams, msg...)
}

//goland:noinspection ALL
func ErrInvalidParamsf(format string, arg ...interface{}) merrors.Error {
	return Errorf(CodeInvalidParams, format, arg...)
}

//goland:noinspection ALL
func ErrInternal(msg ...interface{}) merrors.Error {
	return NewErr(CodeInternalServer, msg...)
}

//goland:noinspection ALL
func ErrInternalf(format string, arg ...interface{}) merrors.Error {
	return Errorf(CodeInternalServer, format, arg...)
}

//endregion

// Errors
//
//goland:noinspection ALL
var (
	Ok          = NewErr(CodeOk)
	ErrNotFound = NewErr(CodeNotFound)
	//ErrDuplicateAdd      = NewErr(CodeDuplicateAdd)
	//ErrInvalidParams     = NewErr(CodeInvalidParams)
	//ErrInternalServer = NewErr(CodeInternalServer)
)
