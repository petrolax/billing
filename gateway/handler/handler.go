package handler

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	tr "github.com/petrolax/billing/transaction"
	"github.com/streadway/amqp"
)

type Handler struct {
	ch  *amqp.Channel
	log *log.Logger
}

func NewHandler(ch *amqp.Channel) *Handler {
	return &Handler{
		ch:  ch,
		log: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func serverResponse(c *gin.Context, code int, message string, result interface{}, err string) {
	c.JSON(code, map[string]interface{}{
		"StatusCode": code,
		"Message":    message,
		"Result":     result,
		"Error":      err,
	})
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(65 + rand.Intn(25))
	}
	return string(bytes)
}

func (h *Handler) Withdraw(c *gin.Context) {
	var transaction tr.Transaction
	if err := c.BindJSON(&transaction); err != nil {
		h.log.Println("Withdraw:Error:", err)
		serverResponse(c, http.StatusInternalServerError, "", "", err.Error())
		return
	}

	q, err := h.ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		h.log.Println("Withdraw:Error:", err)
		serverResponse(c, http.StatusInternalServerError, "Failed to declare a queue", "", err.Error())
		return
	}

	msgs, err := h.ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		h.log.Println("Withdraw:Error:", err.Error())
		serverResponse(c, http.StatusInternalServerError, "Failed to register a consumer", "", err.Error())
		return
	}

	corrId := randomString(32)
	body, _ := json.Marshal(transaction)
	err = h.ch.Publish(
		"",         // exchange
		"withdraw", // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          body,
		})
	if err != nil {
		h.log.Println("Withdraw:Error:", err.Error())
		serverResponse(c, http.StatusInternalServerError, "Failed to publish a message", "", err.Error())
		return
	}

	for d := range msgs {
		if corrId == d.CorrelationId {
			err = json.Unmarshal(d.Body, &transaction)
			// err = json.NewDecoder(d.Body).Decode(&transaction)
			if err != nil {
				h.log.Println("Withdraw:Error:", err.Error())
				serverResponse(c, http.StatusInternalServerError, "Failed to convert body to integer", "", err.Error())
				return
			}
			break
		}
	}

	serverResponse(c, http.StatusOK, "", transaction, "")
}

func (h *Handler) Replenishment(c *gin.Context) {

}

func (h *Handler) CardToCard(c *gin.Context) {

}
