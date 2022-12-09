package gprotocol

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"

	"github.com/golang/protobuf/proto"
)

const (
	MaxCompressSize    = 1024
	CmdHeaderSize      = 8
	MsgErrorHeaderSize = 4
)

var (
	errMsgDataTooShort = errors.New("err msg data too short")
)

// async msg | 3bytes length | 1byte flag | 1byte mainCmd | 3bytes subCmd     | 		data		|
// err msg   | 3bytes length | 1byte flag | error detail |

type Message = proto.Message

/*type Message interface {
	Marshal() (data []byte, err error)
	MarshalTo(data []byte) (n int, err error)
	Size() (n int)
	Unmarshal(data []byte) error
}*/

func zlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	_, err := w.Write(src)
	if err != nil {
		return nil
	}
	w.Close()
	return in.Bytes()
}

func zlibUnCompress(src []byte) []byte {
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

func EncodeMsg(mainCmd uint8, subCmd uint32, msg Message) ([]byte, byte, error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, 0, err
	}
	var (
		mFlag byte
		mBuff []byte
	)
	mFlag = 0 // TODO
	if len(data) >= MaxCompressSize {
		mBuff = zlibCompress(data)
		mFlag |= MsgFlagCompress
	} else {
		mBuff = data
	}
	p := make([]byte, len(mBuff)+CmdHeaderSize)
	p[4] = mainCmd
	p[5] = byte(subCmd >> 16)
	p[6] = byte(subCmd >> 8)
	p[7] = byte(subCmd)
	copy(p[CmdHeaderSize:], mBuff)
	return p, mFlag, nil
}

func DecodeMsg(buf []byte, pb Message) error {
	if len(buf) < CmdHeaderSize {
		return errMsgDataTooShort
	}
	var start = CmdHeaderSize
	flag := buf[3]
	// if flag&MsgFlagUId == MsgFlagUId {
	// 	uid, size := DecodeVarintReverse(buf, len(buf)-1)
	// 	buf = buf[:size]
	// 	userid = uid
	// }

	// if flag&MsgFlagAsync != MsgFlagAsync {
	// 	if len(buf) < CmdHeaderSize+CmdSeqSize {
	// 		glog.Error("[协议] 数据错误 ", buf)
	// 		return
	// 	}
	// 	start += CmdSeqSize
	// }

	var mBuff []byte
	if flag&MsgFlagCompress == MsgFlagCompress {
		mBuff = zlibUnCompress(buf[start:]) // 后续通过接口来来处理,实现可选压缩方式
	} else {
		mBuff = buf[start:]
	}
	err := proto.Unmarshal(mBuff, pb)
	if err != nil {
		return err
	}
	return nil
}

func GetMainCmd(buf []byte) uint8 {
	if len(buf) < CmdHeaderSize {
		return 0
	}
	return buf[4]
}

func GetSubCmd(buf []byte) (cmd uint32) {
	if len(buf) < CmdHeaderSize {
		return
	}
	cmd = uint32(buf[5])<<16 | uint32(buf[6])<<8 | uint32(buf[7]) // ??? 大端，高位在前
	return
}

// func GetSeqId(buf []byte) (seqId uint32) {
// 	if len(buf) < CmdHeaderSize {
// 		return
// 	}
// 	if buf[3]&MsgFlagAsync == 0 {
// 		if len(buf) < CmdHeaderSize+CmdSeqSize {
// 			return
// 		}
// 		seqId = uint32(buf[8])<<24 | uint32(buf[9])<<16 | uint32(buf[10])<<8 | uint32(buf[11])
// 	}
// 	return
// }
//
// func VarintSize(x uint64) (n int) {
// 	for {
// 		n++
// 		x >>= 7
// 		if x == 0 {
// 			break
// 		}
// 	}
// 	return n
// }
//
// func EncodeVarintReverse(data []byte, offset int, v uint64) int {
// 	for v >= 0x80 {
// 		data[offset] = uint8(v&0x7f | 0x80)
// 		v >>= 7
// 		offset--
// 	}
// 	data[offset] = uint8(v)
// 	return offset - 1
// }
//
// func DecodeVarintReverse(data []byte, offset int) (p uint64, newOffset int) {
// 	newOffset = offset
// 	for shift := uint(0); newOffset >= 0; shift += 7 {
// 		b := data[newOffset]
// 		p |= (uint64(b) & 0x7F) << shift
// 		if b < 0x80 {
// 			break
// 		}
// 		newOffset--
// 	}
// 	return
// }
//
// func MarshalUserId(userid uint64, data []byte) []byte {
// 	l := len(data) + VarintSize(userid)
// 	buff := make([]byte, l)
// 	copy(buff[0:], data)
// 	buff[0] = uint8(l >> 16)
// 	buff[1] = uint8(l >> 8)
// 	buff[2] = uint8(l)
// 	buff[3] |= MsgFlagUId
// 	EncodeVarintReverse(buff, l-1, userid)
// 	return buff
// }
//
// func UnmarshalUserId(data []byte) (userid uint64, out []byte, ok bool) {
// 	if data[3]&MsgFlagUId != MsgFlagUId {
// 		return 0, nil, false
// 	}
// 	l := len(data)
// 	var size int
// 	userid, size = DecodeVarintReverse(data, l-1)
// 	l = size
// 	data[0] = uint8(l >> 16)
// 	data[1] = uint8(l >> 8)
// 	data[2] = uint8(l)
// 	data[3] &= ^MsgFlagUId
// 	return userid, data[:l], true
// }
//
// // async msg | 3bytes length | 1byte flag | 1byte service | 3bytes cmd     | 		data		|
// func CreateServiceMsg(service fcmd.Service, uCmd fcmd.UCmd, flag uint8, pbData []byte) []byte {
// 	l := len(pbData) + 8
// 	buff := make([]byte, l)
// 	buff[0] = uint8(l >> 16)
// 	buff[1] = uint8(l >> 8)
// 	buff[2] = uint8(l)
// 	buff[3] = flag
// 	buff[4] = uint8(service)
// 	buff[5] = uint8(uCmd >> 16)
// 	buff[6] = uint8(uCmd >> 8)
// 	buff[7] = uint8(uCmd)
// 	copy(buff[8:], pbData)
// 	return buff
// }

// EncodeError
// err msg   | 3bytes length | 1byte flag | error detail |
func EncodeError(err error) []byte {
	errStr := err.Error()
	errLen := MsgErrorHeaderSize + len(errStr)
	errMsg := make([]byte, errLen)
	errMsg[0] = uint8(errLen >> 16)
	errMsg[1] = uint8(errLen >> 9)
	errMsg[2] = uint8(errLen)
	errMsg[3] = MsgFlagErr
	copy(errMsg[4:], errStr)
	return errMsg
}

func GetError(data []byte) error {
	flag := data[3]
	if flag&MsgFlagErr == 0 {
		return nil
	}
	return errors.New(string(data[4:]))
}
