package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
)

func main() {
	conn,err := redis.Dial("tcp","10.1.0.42:6379")
	if err != nil{
		fmt.Println(err)
		return
	}
	defer conn.Close()
	fmt.Println("连接成功")
	_,err = conn.Do("set","a","1")
	if err != nil {
		fmt.Println(err)
	}
	va,_:= conn.Do("get","a")
	fmt.Println(va)
	//设置过期时间
	_,_ = conn.Do("expire","a",10)
	v,_:= conn.Do("get","a")
	fmt.Println(v)
}
