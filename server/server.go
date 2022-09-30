package server

import (
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"oldjon.com/base/glog"
)

type IServer interface {
	Init() bool
	MainLoop()
	Reload()
	Final() bool
}

type Server struct {
	closed  bool
	Derived IServer
}

func (s *Server) Close() {
	s.closed = true
	return
}

func (s *Server) IsClosed() bool {
	return s.closed
}

func (s *Server) SetCPUNum(num int) {
	if num > 0 {
		runtime.GOMAXPROCS(num)
	} else if num == -1 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	return
}

func (s *Server) Main() bool {
	defer func() {
		s.Derived.Final()
		if err := recover(); err != nil {
			glog.Error("[异常] ", err, "\n", string(debug.Stack()))
		}
		glog.Info("关闭服务器完成")
		glog.Flush()
	}()

	rand.Seed(time.Now().Unix() ^ int64(os.Getpid()))

	if s.Derived == nil {
		glog.Error("[启动] Server Derived为空 ")
		return false
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGPIPE, syscall.SIGHUP)
	go func() {
		for sig := range ch {
			switch sig {
			case syscall.SIGHUP:
				glog.Info("[服务] 收到重新加载配置信号！")
				s.Derived.Reload()
			case syscall.SIGPIPE:
			default:
				s.Close()
			}
			glog.Info("[服务] 收到信号 ", sig)
		}
	}()

	s.SetCPUNum(runtime.NumCPU())

	glog.Info("[启动] 开始初始化")

	if !s.Derived.Init() {
		glog.Error("[启动] 初始化失败")
		return false
	}

	glog.Info("[启动] 初始化成功")

	for !s.IsClosed() {
		s.Derived.MainLoop()
	}

	return true
}
