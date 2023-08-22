package gdb

import (
	"errors"
)

var (
	ErrValue     = errors.New("ERR_VALUE")
	ErrValueType = errors.New("ERR_VALUE_TYPE")
)

var (
	PanicKeyIsMissing             = "db key missing"
	PanicKeyValueCountUnmatched   = "key value count unmatched"
	PanicValueDstNeedBeSlice      = "value dst need be slice"
	PanicValueNeedBeSlice         = "value need be slice"
	PanicValueDstNeedBePointer    = "value dst object need be pointer"
	PanicValueDstNeedBeAllocated  = "value dst object need be pointer allocated"
	PanicFieldValueCountUnmatched = "field value count unmatched"
	PanicFieldsIsMissing          = "field is missing"
	PanicScoreValueCountUnmatched = "score value count unmatched"
	PanicHSetUnsupportedValueType = "hset unsupported value type"
	PanicValueNotNum              = "value is not number"
)
