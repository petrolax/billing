package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/petrolax/billing/gateway/handler"
	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	h := handler.NewHandler(ch)

	router := gin.Default()
	router.PUT("/withdraw", h.Withdraw)
	router.PUT("/replenishment", h.Replenishment)
	router.PUT("/cardtocard", h.CardToCard)

	router.Run(":8080")
}
