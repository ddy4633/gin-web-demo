package dao

import (
	"fmt"
	"gin-web-demo/conf"
	"github.com/gomodule/redigo/redis"
	"time"
)

var Pool redis.Pool

type RedisHandle struct {
}

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
func (r RedisHandle) InsertDate(key, value string) error {
	//获取一个连接
	c := Pool.Get()
	defer c.Close()
	//插入数据
	_, err := c.Do("SET", key, value)
	if !conf.CheckERR(err, "redis Set Value is Failed") {
		return err
	}
	return err
}

//删除Redis数据
func (r RedisHandle) DeleteData(key string) error {
	//获取一个连接
	c := Pool.Get()
	defer c.Close()
	//删除数据
	_, err := c.Do("DELETE", key)
	if !conf.CheckERR(err, "redis Delete Value is Failed") {
		return err
	}
	return nil
}

//设置Key的超时
func (r RedisHandle) InsertTTLData(key, value, ttl, time string) error {
	//获取一个连接
	c := Pool.Get()
	defer c.Close()
	//设置TTL
	_, err := c.Do("SET", key, value, ttl, time)
	if !conf.CheckERR(err, "redis Delete Value is Failed") {
		return err
	}
	return nil
}

//查询Redis的值
func (r RedisHandle) GetDate(key string) string {
	//获取一个连接
	c := Pool.Get()
	defer c.Close()
	//获取value
	value, err := c.Do("GET", key)
	if !conf.CheckERR(err, "redis Delete Value is Failed") {
		return ""
	}
	//数据转换
	va, _ := redis.String(value, err)
	return va
}

//查询失败的集合
func (r RedisHandle) SmeDate() []string {
	//获取一个连接
	c := Pool.Get()
	defer c.Close()
	//设置value
	re, err := c.Do("smembers", "Failed_List")
	if !conf.CheckERR(err, "redis Delete Value is Failed") {
		return []string{}
	}
	//转换成[]string
	st, _ := redis.Strings(re, err)
	return st
}

//设置redis的失败集合
func (r RedisHandle) SaddDate(key string) error {
	//获取一个连接
	c := Pool.Get()
	defer c.Close()
	//设置value
	_, err := c.Do("SADD", "Failed_List", key)
	if !conf.CheckERR(err, "redis Delete Value is Failed") {
		return err
	}
	return nil
}
