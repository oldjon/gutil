package server

import (
	"context"
	"net"
	"time"

	"oldjon.com/base/glog"
)

type TCPGHandler func(ctx context.Context, conn *net.TCPConn)

type TCPServer interface {
	BindAccept(ctx context.Context, address string, handler TCPGHandler) error
}

type tcpServer struct {
	listener *net.TCPListener
}

func (ts *tcpServer) Bind(address string) error {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		glog.Error("[服务] 解析失败 ", address)
		return err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		glog.Error("[服务] 侦听失败 ", address)
		return err
	}

	glog.Info("[服务] 侦听成功 ", address)
	ts.listener = listener
	return nil
}

func (ts *tcpServer) Accept() (*net.TCPConn, error) {

	ts.listener.SetDeadline(time.Now().Add(time.Second * 1))

	conn, err := ts.listener.AcceptTCP()
	if err != nil {
		return nil, err
	}

	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Minute)
	conn.SetNoDelay(true)
	conn.SetWriteBuffer(128 * 1024)
	conn.SetReadBuffer(128 * 1024)

	return conn, nil
}

func (ts *tcpServer) BindAccept(ctx context.Context, address string, handler TCPGHandler) error {
	err := ts.Bind(address)
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := ts.Accept()
			if err != nil {
				continue
			}
			handler(ctx, conn)
		}
	}()
	return nil
}

func (ts *tcpServer) CLose() error {
	return ts.listener.Close()
}

func NewTCPServer(ctx context.Context, address string, handler TCPGHandler) (TCPServer, error) {
	ts := &tcpServer{}
	err := ts.BindAccept(ctx, address, handler)
	if err != nil {
		return nil, err
	}
	return ts, nil
}
