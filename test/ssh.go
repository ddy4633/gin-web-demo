package main

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

func main() {
	password := "Qaz520133~"
	var(
		config *ssh.ClientConfig
		auth []ssh.AuthMethod
	)

	//获取认证方式
	auth = make([]ssh.AuthMethod,0)
	auth = append(auth,ssh.Password(password))
	//创建回调函数
	hostKeyCallbk := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}
	config = &ssh.ClientConfig{
		User:              "dongdy",
		Auth:              auth,
		HostKeyCallback:   hostKeyCallbk,
		BannerCallback:    nil,
		ClientVersion:     "",
		HostKeyAlgorithms: nil,
		Timeout:           5*time.Second,
	}
	//创建连接
	client,err := ssh.Dial("tcp","10.10.8.12:22",config)
	if err != nil {
		fmt.Println("连接异常->",err)
		return
	}
	fmt.Println("ssh.Doal 连接成功!")
	defer client.Close()
	//创建session
	session,err := client.NewSession()
	if err != nil {
		fmt.Println("Create Session failed ->",err)
	}

	err = RunCMD(session)
	fmt.Println("err->",err)
}

func RunCMD(session *ssh.Session)error{
	defer session.Close()
	var stderr,stdin,stdout bytes.Buffer
	var test byte
	session.Stderr = &stderr
	session.Stdin = &stdin
	session.Stdout = &stdout

	err := session.Run("ping -c 2 baidu.com")
	if err != nil {
		//fmt.Println("执行命令 -> ",err)
		return err
	}
	line,err := stdout.ReadString(test)
	if err != nil {
		fmt.Println("标准读取ERR",err)
	}
	fmt.Println("信息",line)
	return nil
}
