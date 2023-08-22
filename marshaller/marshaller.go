package gmarshaller

import (
	"encoding/json"

	"github.com/DataDog/zstd"
	"github.com/golang/protobuf/proto"
)

const (
	MarshallerTypeJSON            = "json"
	MarshallerTypeJSONNum         = 1
	MarshallerTypeProtoBuf        = "protobuf"
	MarshallerTypeProtoBufNum     = 2
	MarshallerTypeProtoBufComp    = "protobufcomp"
	MarshallerTypeProtoBufCompNum = 3
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
	return nil, ErrCanNotMarshalAsProtobufMsg
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
	return ErrCanNotUnmarshalAsProtobufMsg
}

type ProtoCompressMarshaller struct{}

const (
	ProtoCompressLengthLimit = 512
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
		return nil, ErrCanNotMarshalAsProtobufMsg
	}
	if len(data) < ProtoCompressLengthLimit {
		return data, nil
	}

	data, err = zstd.Compress(nil, data)
	if err != nil {
		return nil, ErrCompressFailed
	}
	var newData = make([]byte, len(data)+1)
	copy(newData[1:], data)
	newData[0] = ProtoCompressFlagZstd
	return newData, nil
}

func (pcm *ProtoCompressMarshaller) Unmarshal(data []byte, v interface{}) error {
	if len(data) == 0 {
		return ErrUnmarshalFailed
	}
	if data[0] == ProtoCompressFlagZstd {
		var err error
		data, err = zstd.Decompress(nil, data[1:])
		if err != nil {
			return ErrDecompressFailed
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

	return ErrCanNotUnmarshalAsProtobufMsg
}
