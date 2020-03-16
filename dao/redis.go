package dao

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

var Pool redis.Pool

//建立池连接
func NewRedis() {
	Pool = redis.Pool{
		Dial:            redDial,
		TestOnBorrow:    testBorrow,
		MaxIdle:         16,
		MaxActive:       32,
		IdleTimeout:     60,
		MaxConnLifetime: 3,
	}
}

//每次从池中拿出进行心跳检查
func testBorrow(conn redis.Conn, t time.Time) error {
	//t为当前连接被放回池的时间,当连接放回池1分钟内直接返回
	if time.Since(t) < time.Minute {
		return nil
	}
	//超过了一分钟做ping的操作
	_, err := conn.Do("PING")
	return err
}

//设置建立连接方式
func redDial() (redis.Conn, error) {
	c, err := redis.Dial("tcp", "10.1.0.42:6379",
		redis.DialConnectTimeout(10*time.Second),
		redis.DialReadTimeout(5*time.Second),
		redis.DialWriteTimeout(3*time.Second))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return c, err
}

//插入redis的数据
func InsertDate() {

}
