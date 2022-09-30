package bytebuffer

const (
	preSize  = 0
	initSize = 100
)

type ByteBuffer struct {
	_buffer      []byte
	_prependSize int
	_readerIndex int
	_writerIndex int
}

func NewByteBuffer() *ByteBuffer {
	return &ByteBuffer{
		_buffer:      make([]byte, preSize+initSize),
		_prependSize: preSize,
		_readerIndex: preSize,
		_writerIndex: preSize,
	}
}

func (bb *ByteBuffer) Append(buff ...byte) {
	size := len(buff)
	if size == 0 {
		return
	}
	bb.WriteGrow(size)
	copy(bb._buffer[bb._writerIndex:], buff)
	bb.WriteFlip(size)
}

func (bb *ByteBuffer) WriteBuf() []byte {
	if bb._writerIndex >= len(bb._buffer) {
		return nil
	}
	return bb._buffer[bb._writerIndex:]
}

func (bb *ByteBuffer) WriteSize() int {
	return len(bb._buffer) - bb._writerIndex
}

func (bb *ByteBuffer) WriteFlip(size int) {
	bb._writerIndex += size
}

func (bb *ByteBuffer) WriteGrow(size int) {
	if size > bb.WriteSize() {
		bb.writeReserve(size)
	}
}

func (bb *ByteBuffer) ReadBuf() []byte {
	if bb._readerIndex >= len(bb._buffer) {
		return nil
	}
	return bb._buffer[bb._readerIndex:]
}

func (bb *ByteBuffer) ReadReady() bool {
	return bb._writerIndex > bb._readerIndex
}

func (bb *ByteBuffer) ReadSize() int {
	return bb._writerIndex - bb._readerIndex
}

func (bb *ByteBuffer) RdFlip(size int) {
	if size < bb.ReadSize() {
		bb._readerIndex += size
	} else {
		bb.Reset()
	}
}

func (bb *ByteBuffer) Reset() {
	bb._readerIndex = bb._prependSize
	bb._writerIndex = bb._prependSize
}

func (bb *ByteBuffer) MaxSize() int {
	return len(bb._buffer)
}

func (bb *ByteBuffer) writeReserve(size int) {
	if bb.WriteSize()+bb._readerIndex < size+bb._prependSize {
		tmpBuff := make([]byte, bb._writerIndex+size)
		copy(tmpBuff, bb._buffer)
		bb._buffer = tmpBuff
	} else {
		readable := bb.ReadSize()
		copy(bb._buffer[bb._prependSize:], bb._buffer[bb._readerIndex:bb._writerIndex])
		bb._readerIndex = bb._prependSize
		bb._writerIndex = bb._readerIndex + readable
	}
}

func (bb *ByteBuffer) Prepend(buff []byte) bool {
	size := len(buff)
	if bb._readerIndex < size {
		return false
	}
	bb._readerIndex -= size
	copy(bb._buffer[bb._readerIndex:], buff)
	return true
}
