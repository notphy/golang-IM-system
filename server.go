package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// online User
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 监听Message广播消息channel的goroutine
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	fmt.Println("new server success")
	return server
}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	// ..当前连接的业务
	// fmt.Println("连接建立成功")
	user := NewUser(conn, this)

	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read error:", err)
				return
			}
			// 提取用户的消息
			msg := string(buf[:n-1])

			// 用户针对message进行处理
			user.DoMessage(msg)

			// 任意消息代表用户活跃
			isLive <- true
		}
	}()

	// 当前handler阻塞
	for {
		select {
		case <-isLive:
			// 用户是活跃的，重置定时器
			// 不做任何事情，为了激活select，更新下面的定时器

		case <-time.After(time.Second * 300):
			// 已经超时
			// 将当前的user强制关闭

			user.sendMsg("您已经被踢除")

			// 销毁资源
			close(user.C)

			// 关闭连接
			conn.Close()

			// 退出当前的handler
			return // runtime.exit()
		}
	}

}

// 启动服务器接口
func (this *Server) Start() {
	// socker listen
	fmt.Println("server start")
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go this.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		// do handler
		go this.Handler(conn)
	}

}
