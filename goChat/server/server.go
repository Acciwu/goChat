package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Server类
type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

// Server构造函数
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}


//启动服务器的接口
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//close listen socket
	defer listener.Close()

	//启动监听Message的goroutine
	go this.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		//do handler
		go this.Handler(conn)
	}
}


//监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		//将msg发送给全部的在线User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// go
func (this *Server) Handler(conn net.Conn) {
	//...当前链接的业务
	fmt.Println(conn.RemoteAddr(), "链接建立成功")

	user := NewUser(conn, this)

	// 处理用户上线
	user.OnLine()
	// 上线后，心跳
	actFlagChan := make(chan bool)

	//新开一个go程，接收客户端消息并广播
	go func() {
		buf := make([]byte, 1024*4)

		for  {
			read, err := conn.Read(buf)

			if read == 0{
				user.OffLine()
				return
			}
			if err != nil && err != io.EOF{
				fmt.Println("Conn Read error:", err)
				return
			}

			msg := string(buf[:read-1])

			// 用户针对msg进行处理
			user.DealMsg(msg)
			actFlagChan <- true
		}
	}()
	
	//当前handler阻塞
	for{
		select {
		// 消耗心跳
		case <- actFlagChan:
		// 定时关闭器
		case <- time.After(time.Second * 10-1):
			conn.Write([]byte("您已被踢下线\n"))
			time.Sleep(time.Second * 1)

			conn.Close()
			break
		}
	}
}


//广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}
