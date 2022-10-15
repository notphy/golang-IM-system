package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// 创建用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	// 启动监听当前user channel消息的gorutine
	go user.ListenMessage()

	return user
}

// 用户上线业务
func (this *User) Online() {
	// 用户上线， 将用户加入到onlinemap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线
	this.server.BroadCast(this, "已上线")
}

// 用户下线业务
func (this *User) Offline() {
	// 用户下线， 将用户从onlinemap删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播当前用户下线
	this.server.BroadCast(this, "下线")
}

// 给当前用户对应的客户端发送消息
func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]

		// 判断name是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.sendMsg("当前用户名已被使用！\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()
			this.Name = newName
			this.sendMsg("您已经更新用户名为：" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[0:3] == "to|" {
		// 消息格式：to|张三|消息内容

		// 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.sendMsg("消息格式不正确\n")
			return
		}

		// 根据用户名获取对方的user对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.sendMsg("该用户名不存在\n")
			return
		}

		// 取得消息内容，通过对方的User对象将消息发送
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.sendMsg("请输入消息内容\n")
			return
		} else {
			remoteUser.sendMsg(this.Name + "对您说" + content + "\n")
		}
	}
	this.server.BroadCast(this, msg)
}

// 监听当前User channel，一旦有消息直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		// fmt.Println(msg)
		this.conn.Write([]byte(msg + "\n"))
	}
}
