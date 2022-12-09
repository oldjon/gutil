package gprotocol

const (
	MsgFlagCompress uint8 = 1
	MsgFlagUId      uint8 = 1 << 1
	MsgFlagErr      uint8 = 1 << 2
	MsgFlagAsync    uint8 = 1 << 3
	MsgFlagPush     uint8 = 1 << 4
	MsgFlagAES      uint8 = 1 << 5
)
