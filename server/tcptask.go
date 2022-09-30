package server

import (
	"io"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"oldjon.com/base/bytebuffer"

	"oldjon.com/base/glog"
)

type ITcpTask interface {
	ParseMsg(data []byte) bool
	OnClose()
}

const (
	cmdMaxSize    = 128 * 1024
	cmdHeaderSize = 4 // 3字节指令，1字节flag
	cmdVerifyTime = 30
)

type TCPTask struct {
	closed      int32
	verified    bool
	stoppedChan chan struct{}
	recvBuff    *bytebuffer.ByteBuffer
	sendBuff    *bytebuffer.ByteBuffer
	sendMutex   sync.Mutex
	sendChan    chan struct{}
	Conn        net.Conn
	Derived     ITcpTask
}

func NewTCPTask(conn net.Conn) *TCPTask {
	return &TCPTask{
		closed:      -1,
		verified:    false,
		Conn:        conn,
		stoppedChan: make(chan struct{}, 1),
		recvBuff:    bytebuffer.NewByteBuffer(),
		sendBuff:    bytebuffer.NewByteBuffer(),
		sendChan:    make(chan struct{}, 1),
	}
}

func (tt *TCPTask) SendSignal() {
	select {
	case tt.sendChan <- struct{}{}:
	default:
	}
	return
}

func (tt *TCPTask) RemoteAddr() string {
	if tt.Conn == nil {
		return ""
	}
	return tt.Conn.RemoteAddr().String()
}

func (tt *TCPTask) LocalAddr() string {
	if tt.Conn == nil {
		return ""
	}
	return tt.Conn.LocalAddr().String()
}

func (tt *TCPTask) IsClosed() bool {
	return atomic.LoadInt32(&tt.closed) != 0
}

func (tt *TCPTask) Stop() bool {
	if tt.IsClosed() {
		glog.Error("[连接] 关闭失败 ", tt.RemoteAddr())
		return false
	}
	select {
	case tt.stoppedChan <- struct{}{}:
	default:
		glog.Error("[连接] 关闭失败 ", tt.RemoteAddr())
		return false
	}
	return true
}

func (tt *TCPTask) Start() {
	if !atomic.CompareAndSwapInt32(&tt.closed, -1, 0) {
		return
	}
	job := &sync.WaitGroup{}
	job.Add(1)
	go tt.SendLoop(job)
	go tt.RecvLoop()
	job.Wait()
	glog.Info("[连接] 收到连接 ", tt.RemoteAddr())
	return
}

func (tt *TCPTask) Close() {
	if !atomic.CompareAndSwapInt32(&tt.closed, 0, 1) {
		return
	}
	glog.Info("[连接] 断开连接 ", tt.RemoteAddr())
	tt.Conn.Close()
	tt.recvBuff.Reset()
	tt.sendBuff.Reset()
	select {
	case tt.stoppedChan <- struct{}{}:
	default:
		glog.Error("[连接] 关闭失败 ", tt.RemoteAddr())
	}
	tt.Derived.OnClose()
	return
}

func (tt *TCPTask) Reset() bool {
	if atomic.LoadInt32(&tt.closed) != 1 {
		return false
	}
	if !tt.IsVerified() {
		return false
	}
	tt.closed = -1
	tt.verified = true
	tt.stoppedChan = make(chan struct{})
	glog.Info("[连接] 重置连接 ", tt.RemoteAddr())
	return true
}

func (tt *TCPTask) Verify() {
	tt.verified = true
	return
}

func (tt *TCPTask) IsVerified() bool {
	return tt.verified
}

func (tt *TCPTask) Terminate() {
	tt.Close()
}

func (tt *TCPTask) SendBytes(buffer []byte) bool {
	if tt.IsClosed() {
		return false
	}
	tt.sendMutex.Lock()
	tt.sendBuff.Append(buffer...)
	tt.sendMutex.Unlock()
	tt.SendSignal()
	return true
}

func (tt *TCPTask) readAtLeast(buff *bytebuffer.ByteBuffer, needNum int) error {
	buff.WriteGrow(needNum)
	n, err := io.ReadAtLeast(tt.Conn, buff.WriteBuf(), needNum)
	buff.WriteFlip(n)
	return err
}

func (tt *TCPTask) RecvLoop() {
	defer func() {
		tt.Close()
		if err := recover(); err != nil {
			glog.Error("[异常] ", err, "\n", string(debug.Stack()))
		}
	}()

	var (
		needNum   int
		err       error
		totalSize int
		dataSize  int
		msgBuff   []byte
	)

	for {

		totalSize = tt.recvBuff.ReadSize()
		if totalSize <= cmdHeaderSize {
			needNum = cmdHeaderSize - totalSize
			err = tt.readAtLeast(tt.recvBuff, needNum)
			if err != nil {
				glog.Error("[连接] 接收数据失败 ", tt.RemoteAddr(), ",", err)
				return
			}
			totalSize = tt.recvBuff.ReadSize()
		}

		msgBuff = tt.recvBuff.ReadBuf()

		dataSize = (int(msgBuff[0]) << 16) | (int(msgBuff[1]) << 8) | int(msgBuff[2])
		if dataSize > cmdMaxSize {
			glog.Error("[连接] 数据长度超过最大值 ", tt.RemoteAddr(), ",", dataSize)
			return
		} else if dataSize < cmdHeaderSize {
			glog.Error("[连接] 数据长度不足最小值 ", tt.RemoteAddr(), ",", dataSize)
			return
		}

		if totalSize < dataSize {
			needNum = dataSize - totalSize
			err = tt.readAtLeast(tt.recvBuff, needNum)
			if err != nil {
				glog.Error("[连接] 接收数据失败 ", tt.RemoteAddr(), ",", err)
				return
			}
			msgBuff = tt.recvBuff.ReadBuf()
		}

		tt.Derived.ParseMsg(msgBuff[:dataSize])
		tt.recvBuff.RdFlip(dataSize)
	}
}

func (tt *TCPTask) SendLoop(job *sync.WaitGroup) {
	defer func() {
		tt.Close()
		if err := recover(); err != nil {
			glog.Error("[异常] ", err, "\n", string(debug.Stack()))
		}
	}()

	var (
		tmpByte  = bytebuffer.NewByteBuffer()
		timeout  = time.NewTimer(time.Second * cmdVerifyTime)
		writeNum int
		err      error
	)

	defer timeout.Stop()

	job.Done()

	for {
		select {
		case <-tt.sendChan:
			for {
				tt.sendMutex.Lock()
				if tt.sendBuff.ReadReady() {
					tmpByte.Append(tt.sendBuff.ReadBuf()[:tt.sendBuff.ReadSize()]...)
					tt.sendBuff.Reset()
				}
				tt.sendMutex.Unlock()

				if !tmpByte.ReadReady() {
					break
				}

				writeNum, err = tt.Conn.Write(tmpByte.ReadBuf()[:tmpByte.ReadSize()])
				if err != nil {
					glog.Error("[连接] 发送失败 ", tt.RemoteAddr(), ",", err)
					return
				}
				tmpByte.RdFlip(writeNum)
			}
		case <-tt.stoppedChan:
			return
		case <-timeout.C:
			if !tt.IsVerified() {
				glog.Error("[连接] 验证超时 ", tt.RemoteAddr())
				return
			}
		}
	}
}
