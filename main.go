package main

import (
	"github.com/26tanishabanik/customController/redis"
	"github.com/26tanishabanik/customController/consumer"
	"github.com/26tanishabanik/customController/pods"
)

func main() {
	redis.Init()
	go consumer.InitConsumer()
	go pods.Both()

}

