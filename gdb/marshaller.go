package gdb

import (
	"encoding/json"
	"errors"

	"github.com/DataDog/zstd"
	"github.com/golang/protobuf/proto"
)

var (
	ErrCanNotMarshalAsProtobufMessage   = errors.New("ErrCanNotMarshalAsProtobufMessage")
	ErrCanNotUnmarshalAsProtobufMessage = errors.New("ErrCanNotUnmarshalAsProtobufMessage")
	ErrCompressBytesFailed              = errors.New("ErrCompressBytesFailed")
	ErrDecompressBytesFailed            = errors.New("ErrDecompressBytesFailed")
)

type Marshaller interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

type JsonMarshaller struct{}

func (jm *JsonMarshaller) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jm *JsonMarshaller) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type ProtoMarshaller struct{}

type SelfMarshaller interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

func (pm *ProtoMarshaller) Marshal(v interface{}) ([]byte, error) {
	ms, ok := v.(SelfMarshaller)
	if ok {
		return ms.Marshal()
	}
	mp, ok := v.(proto.Message)
	if ok {
		return proto.Marshal(mp)
	}
	return nil, ErrCanNotMarshalAsProtobufMessage
}

func (pm *ProtoMarshaller) Unmarshal(data []byte, v interface{}) error {
	ms, ok := v.(SelfMarshaller)
	if ok {
		return ms.Unmarshal(data)
	}
	mp, ok := v.(proto.Message)
	if ok {
		return proto.Unmarshal(data, mp)
	}
	return ErrCanNotUnmarshalAsProtobufMessage
}

type ProtoCompressMarshaller struct{}

const (
	ProtoCompressLengthLimit = 1024
	ProtoCompressFlagZstd    = byte(1)<<6 | 7 // 01000111
)

func (pcm *ProtoCompressMarshaller) Marshal(v interface{}) ([]byte, error) {
	var data []byte
	var err error
	ms, ok := v.(SelfMarshaller)
	if ok {
		data, err = ms.Marshal()
	} else {
		mp, ok := v.(proto.Message)
		if ok {
			data, err = proto.Marshal(mp)
		}
	}
	if err != nil {
		return nil, ErrCanNotMarshalAsProtobufMessage
	}
	if len(data) < ProtoCompressLengthLimit {
		return data, nil
	}

	data, err = zstd.Compress(nil, data)
	var newData = make([]byte, len(data)+1)
	copy(newData[1:], data)
	newData[0] = ProtoCompressFlagZstd
	return newData, nil
}

func (pcm *ProtoCompressMarshaller) Unmarshal(data []byte, v interface{}) error {
	if len(data) > 0 && data[0] == ProtoCompressFlagZstd {
		var err error
		data, err = zstd.Decompress(nil, data[1:])
		if err != nil {
			return ErrDecompressBytesFailed
		}
	}
	ms, ok := v.(SelfMarshaller)
	if ok {
		return ms.Unmarshal(data)
	}
	mp, ok := v.(proto.Message)
	if ok {
		return proto.Unmarshal(data, mp)
	}

	return ErrCanNotUnmarshalAsProtobufMessage
}
