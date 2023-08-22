package gmarshaller

import "errors"

var (
	ErrCanNotMarshalAsProtobufMsg   = errors.New("ERR_CAN_NOT_MARSHAL_AS_PROTOBUF_MSG")
	ErrCanNotUnmarshalAsProtobufMsg = errors.New("ERR_CAN_NOT_UNMARSHAL_AS_PROTOBUF_MSG")
	ErrCompressFailed               = errors.New("ERR_COMPRESS_FAILED")
	ErrDecompressFailed             = errors.New("ERR_DECOMPRESS_FAILED")
	ErrMarshalFailed                = errors.New("ERR_DB_MARSHAL_FAILED")
	ErrUnmarshalFailed              = errors.New("ERR_DB_UNMARSHAL_FAILED")
)
