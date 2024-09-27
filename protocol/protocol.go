package gprotocol

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"

	gmarshaller "github.com/oldjon/gutil/marshaller"
)

const (
	MaxCompressSize    = 1024
	HeaderSize         = 4
	CmdHeaderSize      = 8
	MsgErrorHeaderSize = 4
)

var (
	ErrMsgDataTooShort = errors.New("err msg data too short")
)

// async msg [ 3bytes length | 1byte flag | 1byte mainCmd | 3bytes subCmd     | 		data		]

type FrameCoder interface {
	EncodeMsg(mainCmd uint8, subCmd uint32, msg interface{}) ([]byte, error)
	DecodeMsg(buf []byte, pb interface{}) error
	MainCmd(buf []byte) uint8
	SubCmd(buf []byte) (cmd uint32)
	HeaderSize() int
	Size(data []byte) (int, error)
}

type frameCoder struct {
	MarshallerType uint8
	Marshaller     gmarshaller.Marshaller
}

func newFrameCoder(marshallType string) *frameCoder {
	ff := &frameCoder{}
	switch marshallType {
	case gmarshaller.MarshallerTypeJSON:
		ff.Marshaller = &gmarshaller.JsonMarshaller{}
		ff.MarshallerType = gmarshaller.MarshallerTypeJSONNum
	case gmarshaller.MarshallerTypeProtoBuf:
		ff.Marshaller = &gmarshaller.ProtoMarshaller{}
		ff.MarshallerType = gmarshaller.MarshallerTypeProtoBufNum
	case gmarshaller.MarshallerTypeProtoBufComp: // will still use ProtoMarshaller without compress
		ff.Marshaller = &gmarshaller.ProtoMarshaller{}
		ff.MarshallerType = gmarshaller.MarshallerTypeProtoBufNum
	default:
		ff.Marshaller = &gmarshaller.JsonMarshaller{}
		ff.MarshallerType = gmarshaller.MarshallerTypeJSONNum
	}
	return ff
}

func NewFrameCoder(marshallType string) FrameCoder {
	return newFrameCoder(marshallType)
}

func (fc *frameCoder) zlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	_, err := w.Write(src)
	if err != nil {
		return nil
	}
	w.Close()
	return in.Bytes()
}

func (fc *frameCoder) zlibUnCompress(src []byte) []byte {
	b := bytes.NewReader(src)
	var out bytes.Buffer
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil
	}
	_, err = io.Copy(&out, r)
	if err != nil {
		return nil
	}
	return out.Bytes()
}

func (fc *frameCoder) EncodeMsg(mainCmd uint8, subCmd uint32, msg interface{}) ([]byte, error) {
	data, err := fc.Marshaller.Marshal(msg)
	if err != nil {
		return nil, err
	}
	var (
		mFlag byte
		mBuff []byte
	)
	if fc.MarshallerType == gmarshaller.MarshallerTypeJSONNum {
		mFlag |= MsgFlagJSON
	} else if fc.MarshallerType == gmarshaller.MarshallerTypeProtoBufNum {
		mFlag |= MsgFlagProtoBuf
	}

	if fc.MarshallerType != gmarshaller.MarshallerTypeProtoBufCompNum &&
		len(data) >= MaxCompressSize {
		mBuff = fc.zlibCompress(data)
		mFlag |= MsgFlagCompress
	} else {
		mBuff = data
	}
	size := len(mBuff) + CmdHeaderSize
	p := make([]byte, size)
	p[0] = uint8(size >> 16)
	p[1] = uint8(size >> 8)
	p[2] = uint8(size)
	p[3] = mFlag
	p[4] = mainCmd
	p[5] = byte(subCmd >> 16)
	p[6] = byte(subCmd >> 8)
	p[7] = byte(subCmd)
	copy(p[CmdHeaderSize:], mBuff)
	return p, nil
}

func (fc *frameCoder) DecodeMsg(buf []byte, pb interface{}) error {
	if len(buf) < CmdHeaderSize {
		return ErrMsgDataTooShort
	}
	var start = CmdHeaderSize
	flag := buf[3]

	var mBuff []byte
	if flag&MsgFlagCompress == MsgFlagCompress {
		mBuff = fc.zlibUnCompress(buf[start:]) // 后续通过接口来来处理,实现可选压缩方式
	} else {
		mBuff = buf[start:]
	}
	err := fc.Marshaller.Unmarshal(mBuff, pb)
	if err != nil {
		return err
	}
	return nil
}

func (fc *frameCoder) MainCmd(buf []byte) uint8 {
	if len(buf) < CmdHeaderSize {
		return 0
	}
	return buf[4]
}

func (fc *frameCoder) SubCmd(buf []byte) (cmd uint32) {
	if len(buf) < CmdHeaderSize {
		return
	}
	cmd = uint32(buf[5])<<16 | uint32(buf[6])<<8 | uint32(buf[7]) // ??? 大端，高位在前
	return
}

func (fc *frameCoder) data(buf []byte) []byte {
	if len(buf) < CmdHeaderSize {
		return nil
	}
	return buf[CmdHeaderSize:]
}

func (fc *frameCoder) HeaderSize() int {
	return HeaderSize
}

func (fc *frameCoder) Size(data []byte) (int, error) {
	if len(data) < HeaderSize {
		return 0, ErrMsgDataTooShort
	}
	return int(uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])), nil
}

func DecodeMsg(buf []byte, pb interface{}) error {
	if len(buf) < CmdHeaderSize {
		return ErrMsgDataTooShort
	}
	var start = CmdHeaderSize
	flag := buf[3]

	var fc *frameCoder
	if flag&MsgFlagJSON > 0 {
		fc = newFrameCoder(gmarshaller.MarshallerTypeJSON)
	} else if flag&MsgFlagProtoBuf > 0 {
		fc = newFrameCoder(gmarshaller.MarshallerTypeProtoBuf)
	} else {
		fc = newFrameCoder(gmarshaller.MarshallerTypeJSON)
	}

	var mBuff []byte
	if flag&MsgFlagCompress == MsgFlagCompress {
		mBuff = fc.zlibUnCompress(buf[start:]) // 后续通过接口来来处理,实现可选压缩方式
	} else {
		mBuff = buf[start:]
	}
	err := fc.Marshaller.Unmarshal(mBuff, pb)
	if err != nil {
		return err
	}
	return nil
}
