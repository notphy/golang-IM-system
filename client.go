package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int // 当前客户端的模式
}

// 创建客户端对象
func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))

	if err != nil {
		fmt.Println("net Dial error", err)
		return nil
	}

	fmt.Printf("%s:%d", serverIp, serverPort)

	client.conn = conn

	// 返回对象
	return client
}

// 处理server回应的消息，直接显示
func (this *Client) DealResponse() {
	// 一旦conn有数据，就直接copy到stdout白标准输出，永久阻塞监听
	io.Copy(os.Stdout, this.conn)
	// 等价于
	// for {
	// 	buf := make([]byte, 4096)
	// 	this.conn.Read(buf)
	// 	fmt.Println(buf)
	// }
}

func (this *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新同户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		this.flag = flag
		return true
	} else {
		fmt.Println("请输入合法范围的数字\n")
		return false
	}
}

func (this *Client) PublicChat() {
	var chatMsg string
	fmt.Println("请输入聊天内容, exit退出.")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		// 发送给客户端
		// 消息不为空发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write error:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println("请输入聊天内容, exit退出.")
		fmt.Scanln(&chatMsg)
	}
}

func (this *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write error:", err)
		return
	}
}

func (this *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	this.SelectUsers()
	fmt.Println("请输入聊天对象用户名, exit退出.")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println("请输入聊天内容, exit退出.")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			// 发送给客户端
			// 消息不为空发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := this.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write error:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println("请输入聊天内容, exit退出.")
			fmt.Scanln(&chatMsg)
		}

		this.SelectUsers()
		fmt.Println("请输入聊天对象用户名, exit退出.")
		fmt.Scanln(&remoteName)
	}
}

func (this *Client) UpdateName() bool {
	fmt.Println(">>>>>>请输入用户名")
	fmt.Scanln(&this.Name)

	sendMsg := "rename|" + this.Name + "\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return false
	}

	return true
}

func (this *Client) Run() {
	for this.flag != 0 {
		for this.menu() != true {

		}

		// 根据不同的模式处理不同的业务
		switch this.flag {
		case 1:
			// 公聊模式
			this.PublicChat()
			break
		case 2:
			// 私聊模式
			this.PrivateChat()
			break
		case 3:
			// 更新用户名
			this.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

// client.exe -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器默认地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器默认地址(默认是8888)")
}

func main() {
	// 命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>连接服务器失败")
		return
	}

	// 单独开启一个goroutine去处理server的回执消息
	go client.DealResponse()

	fmt.Println(">>>>>>连接服务器成功")

	// 启动客户端业务
	client.Run()
}
