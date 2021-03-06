package dao

import (
	"errors"
	"fmt"
	"gin-web-demo/conf"
	"gin-web-demo/tools"
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
	c, err := redis.Dial("tcp", conf.Config.Conf.DataBase[0].Host,
		redis.DialConnectTimeout(10*time.Second),
		redis.DialReadTimeout(5*time.Second),
		redis.DialWriteTimeout(3*time.Second))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//auth认证
	if _, authErr := c.Do("AUTH", conf.Config.Conf.DataBase[0].Password); authErr != nil {
		return nil, fmt.Errorf("redis auth password error: %s", authErr)
	}
	return c, err
}

//插入redis的数据
func (r RedisHandle) InsertDate(key, value string) error {
	//获取一个连接
	c := Pool.Get()
	defer c.Close()
	//插入数据
	_, err := c.Do("set", key, value)
	if !tools.CheckERR(err, "redis Set Value is Failed") {
		return err
	}
	if len(value) < 21 {
		//插入有序类型
		err = r.SaddDate(key)
		if !tools.CheckERR(err, "redis Set SaddDate Value is Failed") {
			return err
		}
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
	if !tools.CheckERR(err, "redis Delete Value is Failed") {
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
	if !tools.CheckERR(err, "redis Delete Value is Failed") {
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
	if !tools.CheckERR(err, "redis Delete Value is Failed") {
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
	if !tools.CheckERR(err, "redis Delete Value is Failed") {
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
	//获取当前的位置
	num := r.ZcardDate(key)
	//设置value
	_, err := c.Do("ZADD", "FailedList", num, key)
	if !tools.CheckERR(err, "redis Set SaddDate Value is Failed") {
		return err
	}
	return nil
}

//查询有序集合总数
func (r RedisHandle) ZcardDate(key string) int {
	//获取一个连接
	c := Pool.Get()
	defer c.Close()
	//设置value
	v, err := c.Do("ZCARD", "FailedList")
	if !tools.CheckERR(err, "redis query ZCARD Value is Failed") {
		return 0
	}
	num, _ := redis.Int(v, err)
	return num + 1
}

//查询指定范围的有序集合
func (r RedisHandle) ZrangeDate(key string, range1, range2 int) (t []string, err error) {
	//获取一个连接
	c := Pool.Get()
	defer c.Close()
	//查询key
	info, err := c.Do("ZRANGE", key, range1, range2)
	if !tools.CheckERR(err, "redis query ZRANGE Value is Failed") {
		return t, err
	}
	//处理返回的列表
	t, err = redis.Strings(info, err)
	if !tools.CheckERR(err, "redis Delete Value is Failed") {
		return t, err
	}
	return t, nil
}

//删除有序集合中的元素
func (r RedisHandle) ZremDate(key, Member string) error {
	//获取一个连接
	c := Pool.Get()
	defer c.Close()
	//查询key
	_, err := c.Do("SREM", key, Member)
	if !tools.CheckERR(err, "redis ZREM Value is Failed") {
		return err
	}
	return err
}

//任务队列(add集合)
func (r RedisHandle) SaddQueue(key, Member string) error {
	//获取一个连接
	c := Pool.Get()
	defer c.Close()
	//新增member
	a, err := c.Do("SADD", key, Member)
	va, _ := redis.Int(a, err)
	if va == 0 {
		return errors.New("key is active!")
	} else if va != 0 && va != 1 {
		return errors.New("redis zadd is failed,An unknown error has occurred")
	}
	return nil
}
