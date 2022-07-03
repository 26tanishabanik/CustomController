package redis

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Logs struct {
	Description string `redis:"description"`
}

func Init() {
	Pool = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
}

var Pool *redis.Pool


func Store(word string) {
	
	conn := Pool.Get()

	fmt.Println("Connected to redis cache ")
	_, err := conn.Do(
		"HMSET",
		"logs:1",
		"description",
		word,
	)
	if err != nil {
		fmt.Printf("Couldn't set record: %s", err.Error())
	}
	values, err := redis.StringMap(conn.Do("HGETALL", "logs:1"))
	if err != nil{
		fmt.Printf("Couldn't get all records: %s", err.Error())
	}
	for k, v := range values {
		fmt.Println("Redis Key:", k)
		fmt.Println("Redis Value:", v)
	}
	defer conn.Close()
}

