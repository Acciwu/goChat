package main

import (
	"net"
	"strings"
)

// User类
type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// User有参构造
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
		server: server,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

//监听当前User channel的 方法,一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线业务
func (this *User) OnLine()  {
	//用户上线,将用户加入到onlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}
// 用户下线业务
func (this *User) OffLine()  {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "已下线")
}
// 用户处理消息
func (this *User) DealMsg(msg string) {

	if msg == "who"{
		for _,user := range this.server.OnlineMap{
			onLineMsg := "[" + user.Addr  + "]" + ":" + user.Name + "在线中...\n"
			this.conn.Write([]byte(onLineMsg))
		}
	}else if len(msg) > 7 &&  msg[:7] == "rename|" {
		// 获取新用户名
		newName := strings.Split(msg, "|")[1]
		// 修改客户端连接用户名
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.conn.Write([]byte("当前用户名已经存在\n"))
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.conn.Write([]byte("用户名已更新为：" + newName + "\n"))
		}
	}else if len(msg) > 4 && msg[0:3] == "to|" {
			// 获取用户名
			userName := strings.Split(msg, "|")[1]
			if(userName == ""){
				this.conn.Write([]byte("消息格式不正确\n"))
				return
			}

			// 判断用户名是否存在
			toUser, ok := this.server.OnlineMap[userName]
			if !ok {
				this.conn.Write([]byte("用户名[" + userName + "]不存在\n"))
				return
			}

			// 获取要发送的内容
			sendMsg := strings.Split(msg, "|")[2]
			if sendMsg == ""{
				this.conn.Write([]byte("不能发送空消息\n"))
				return
			}

			// 私发消息
			toUser.conn.Write([]byte("from " + this.Name + ": " + sendMsg + "\n"))
		}else {
			this.server.BroadCast(this, msg)
		}
}
