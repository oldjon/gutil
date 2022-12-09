package gredis

import (
	"encoding/json"
	"errors"

	"github.com/golang/protobuf/proto"
)

var (
	ErrCanNotMarshalAsProtobufMessage   = errors.New("ErrCanNotMarshalAsProtobufMessage")
	ErrCanNotUnmarshalAsProtobufMessage = errors.New("ErrCanNotUnmarshalAsProtobufMessage")
)

type ObjMarshaller interface {
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
