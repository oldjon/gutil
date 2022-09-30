package ringbuffer

type Ringbuffer struct {
	start, use int
	buf        []byte
}

func IntToByte(n int) []byte {
	buf := make([]byte, 4)
	buf[3] = (byte)((n >> 24) & 0xFF)
	buf[2] = (byte)((n >> 16) & 0xFF)
	buf[1] = (byte)((n >> 8) & 0xFF)
	buf[0] = (byte)(n & 0xFF)
	return buf
}
func ByteToInt(buf []byte) int {
	var value int
	value = int(buf[0]&0xFF) | (int(buf[1]&0xFF) << 8) | (int(buf[2]&0xFF) << 16) | (int(buf[3]&0xFF) << 24)
	return value
}

func NewRingbuffer(size int) *Ringbuffer {
	return &Ringbuffer{0, 0, make([]byte, size)}
}

// WriteCover 覆盖未读数据包策略
func (r *Ringbuffer) WriteCover(b []byte) bool {
	blockSize := len(b)
	if blockSize > 0 {
		size := len(r.buf)
		start := (r.start + r.use) % size
		sizeByte := IntToByte(blockSize)
		/*判断是非会覆盖未读block，
		  是的话，修改r.start */
		flag := blockSize + len(sizeByte)
		for flag > (r.start-start+size)%size && r.use != 0 {
			rBlockSize := ByteToInt(r.buf[r.start : r.start+4])
			r.start = (r.start + rBlockSize + 4) % size
		}
		// 保存block的长度
		n := copy(r.buf[start:], sizeByte)
		if start+len(sizeByte) > len(r.buf) {
			copy(r.buf, sizeByte[n:]) // 判断是否需要绕回
		}
		start = (start + len(sizeByte)) % size
		// 保存block的内容
		n = copy(r.buf[start:], b)
		if start+len(b) > len(r.buf) {
			copy(r.buf, b[n:]) // 判断是非需要绕回
		}
		start = (start + blockSize) % size
		// 更新ringbuffer的使用量
		r.use = (start + size - r.start) % size
		return true
	}
	return false
}

// Write 丢弃新写入策略
func (r *Ringbuffer) Write(b []byte) bool {
	blockSize := len(b)
	if blockSize > 0 {
		size := len(r.buf)
		start := (r.start + r.use) % size
		sizeByte := IntToByte(blockSize)
		// 判断ringbuffer是否还有空间存放block
		end := start + len(b) + len(sizeByte)
		flag := end - len(r.buf)
		if flag > 0 && flag > r.start {
			return false
		}
		// 保存block的长度
		n := copy(r.buf[start:], sizeByte)
		if start+len(sizeByte) > len(r.buf) {
			copy(r.buf, sizeByte[n:])
		}
		start = (start + len(sizeByte)) % size
		// 保存block的内容
		n = copy(r.buf[start:], b)
		if start+len(b) > len(r.buf) {
			copy(r.buf, b[n:])
		}
		start = (start + blockSize) % size
		// 更新ringbuffer的使用量
		r.use = (start + size - r.start) % size
		return true
	}
	return false
}

func (r *Ringbuffer) Read(b []byte) int {
	if r.use > 0 { // 判断是非还有未读数据
		// 获取block的长度
		sizeByte := make([]byte, 4)
		t := copy(sizeByte, r.buf[r.start:])
		if t != 4 { // 判断有没有被分割开
			copy(sizeByte[t:4], r.buf[:])
		}
		rBlockSize := ByteToInt(sizeByte)
		// 获取block的内容
		start := (r.start + 4) % len(r.buf)
		nRead := 0
		if start+rBlockSize >= len(r.buf) {
			n := copy(b, r.buf[start:]) // 判断数据包内容有没有被分割
			nRead = copy(b[n:], r.buf[:])
			nRead = nRead + n
		} else {
			nRead = copy(b, r.buf[start:start+rBlockSize])
		}
		if nRead == rBlockSize {
			r.start = (r.start + rBlockSize + 4) % len(r.buf)
			r.use = r.use - rBlockSize - 4
			return nRead
		} else {
			return -1
		}
	}
	return 0
}

func (r *Ringbuffer) GetUse() int {
	return r.use
}
