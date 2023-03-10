package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

// Client类
type Client struct {
	ServerIp string
	ServerPort int
	Name string
	conn net.Conn

	flag int
}

// Client类的构造对象
func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp: serverIp,
		ServerPort: serverPort,
		flag: -1,
	}

	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil{
		fmt.Println("net.Dial error", err)
		return nil
	}
	client.conn = conn

	return client
}

// 启动界面
func (this *Client) Run(){
	for this.flag != 0 {
		for this.Menu() != true {
		}

		switch this.flag {
		case 1:
			this.PublicChat()
		case 2:
			this.PrivateChat()
		case 3:
			this.UpdateName()
		}
	}
}

// 菜单选择界面
func (this *Client) Menu() bool{

	var input int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&input)
	if input >= 0 && input <= 3{
		this.flag = input
		return true
	}else{
		fmt.Println(">>>>>>请输入合法数字<<<<<<<")
		return false
	}
}

// 公聊
func (this *Client) PublicChat() {
	this.ShowOnlineUsers()
	fmt.Println(">>>>进入公聊模式, exit退出")
	var chatMsg string
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0{
			sendMsg := chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil{
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>进入公聊模式, exit退出")
		fmt.Scanln(&chatMsg)
	}
}


// 私聊
func (this *Client) PrivateChat()  {
	this.ShowOnlineUsers()
	fmt.Println(">>>>进入私聊模式, exit退出")
	var chatMsg string
	fmt.Scanln(&chatMsg) //

	for chatMsg != "exit" {
		if len(chatMsg) != 0{
			sendMsg := "to|" + chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil{
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>进入私聊模式, exit退出")
		fmt.Scanln(&chatMsg)
	}
}


// 更新用户名
func (this *Client) UpdateName() bool{
	fmt.Print(">>>>请输入新的用户名：")
	// 将输入的新用户名赋值给client.Name
	fmt.Scanln(&this.Name)

	// 像服务器端发送请求，请求更改用户名: user.DealMsg(sendMsg)
	sendMsg := "rename|" + this.Name + "\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil{
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

// 展示在线用户
func (this *Client) ShowOnlineUsers() {
	fmt.Println("在线用户列表>>>>>>>>>")
	_, err := this.conn.Write([]byte("who\n"))
	if err != nil{
		fmt.Println("conn.Write err:", err)
		return
	}
}


// 响应server返回的消息，显示到前端
func (this *Client) DealResponse() {
	io.Copy(os.Stdout, this.conn)
}


// 初始化，加载命令行参数
var serverIp string
var serverPort int
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "默认ip为127.0.0.1")
	flag.IntVar(&serverPort, "port", 8888, "默认端口号为8888")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>连接服务器失败...")
		return
	}

	// 单独开启一个go去处理server的返回信息
	go client.DealResponse()

	fmt.Println(">>>>连接服务器成功...")

	// 客户端操作
	client.Run()




}



