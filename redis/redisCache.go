package redis

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	// "github.com/go-redis/redis"
	// "html/template"
	// "net/http"
	// "strconv"
	// "context"
)

type Logs struct {
	Description string `redis:"description"`
}
// var redisClient *redis.Client
func Init() {
	Pool = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}
	// client := redis.NewClient(&redis.Options{
	// 	Addr: "localhost:6379",
	// 	Password: "",
	// 	DB: 0,
	// })

	// pong, err := client.Ping().Result()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(pong)

	// redisClient = client
}

var Pool *redis.Pool
// var conn redis.Conn = Pool.Get()


func Store(word string) {
	// conn, err := redis.Dial("tcp", "localhost:6379")
	// if err != nil {
	// 	fmt.Printf("Couldn't connect to redis cache: %s", err.Error())
	// }
	conn := Pool.Get()

	fmt.Println("Connected to redis cache ")
	_, err := conn.Do(
		"HMSET",
		"logs:1",
		"description",
		word,
	)
	// err := redisClient.Set("description", word, 0).Err()
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

	// iter := redisClient.Scan(0, "description", 0).Iterator()
	// // keys := []string{}
	// for iter.Next() {
	// 	fmt.Println("Redis Value:", iter.Val())
	// 	// keys = append(keys, iter.Val())
	// }
	// if err := iter.Err(); err != nil {
	// 	fmt.Printf("Couldn't get all records: %s", err.Error())
	// }

	// description, err := redis.String(conn.Do("HGET", "logs:1", "description"))
	// if err != nil{
	// 	fmt.Printf("Couldn't read record: %s", err.Error())
	// }
	// fmt.Printf("Log description is %s", description)
	defer conn.Close()
}
// func PrintCache() []string{
// 	iter := redisClient.Scan(0, "description", 0).Iterator()
// 	keys := []string{}
// 	for iter.Next() {
// 		fmt.Println("Redis Value:", iter.Val())
// 		keys = append(keys, iter.Val())
// 	}
// 	if err := iter.Err(); err != nil {
// 		fmt.Printf("Couldn't get all records: %s", err.Error())
// 	}
// 	return keys
// }

/*
func PrintCache() *Logs{
	// conn, err := redis.Dial("tcp", "localhost:6379")
	// if err != nil {
	// 	fmt.Printf("Couldn't connect to redis cache: %s", err.Error())
	// }
	conn := Pool.Get()
	fmt.Println("Connected to redis cache ")
	values, err := redis.Values(conn.Do("HGETALL", "logs:1"))
	if err != nil {
		fmt.Printf("Couldn't get all records: %s", err.Error())
	}
	// for k, v := range values {
	// 	fmt.Println("Redis Key:", k)
	// 	fmt.Println("Redis Value:", v)
	// }
	var log Logs
	err = redis.ScanStruct(values, &log)
	fmt.Println(log)
	if err != nil {
		fmt.Printf("Couldn't marshal to Logs struct: %s", err.Error())
	}
	return &log
}
*/