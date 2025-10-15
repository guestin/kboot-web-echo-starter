package kerrors

import "github.com/guestin/mob/merrors"

func InternalErr(msg ...interface{}) merrors.Error {
	return NewErr(CodeInternalServer, msg...)
}

//goland:noinspection ALL
func InternalErrf(format string, arg ...interface{}) merrors.Error {
	return Errorf(CodeInternalServer, format, arg...)
}

//goland:noinspection ALL
func DBRecordCreateErr(msg ...interface{}) merrors.Error {
	return NewErr(CodeRecordCreateErr, msg...)
}

//goland:noinspection ALL
func DBRecordUpdateErr(msg ...interface{}) merrors.Error {
	return NewErr(CodeRecordUpdateErr, msg...)
}

//goland:noinspection ALL
func DBRecordRetrieveErr(msg ...interface{}) merrors.Error {
	return NewErr(CodeRecordRetrieveErr, msg...)
}

//goland:noinspection ALL
func DBRecordDeleteErr(msg ...interface{}) merrors.Error {
	return NewErr(CodeRecordDeleteErr, msg...)
}
