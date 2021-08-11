package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	tr "github.com/petrolax/billing/transaction"
	"github.com/streadway/amqp"
	"github.com/teris-io/shortid"
)

type Logger interface {
	Print(v ...interface{})
	Println(v ...interface{})
	Printf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalln(v ...interface{})
	Fatalf(format string, v ...interface{})
	Panic(v ...interface{})
	Panicln(v ...interface{})
	Panicf(format string, v ...interface{})
	Output(calldepth int, s string) error
}

type Handler struct {
	ch    *amqp.Channel
	log   *log.Logger
	queue [3]amqp.Queue
	msgs  [3]<-chan amqp.Delivery
}

func NewHandler(ch *amqp.Channel) *Handler {
	var queue [3]amqp.Queue
	var err error
	for i := 0; i < len(queue); i++ {
		queue[i], err = ch.QueueDeclare(
			"",
			false,
			false,
			true,
			false,
			nil,
		)
		if err != nil {
			log.Fatalln("Withdraw:Error:", err)
		}
	}

	var msgs [3]<-chan amqp.Delivery
	for i := 0; i < len(msgs); i++ {
		msgs[i], err = ch.Consume(
			queue[i].Name,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Fatalln("Withdraw:Error:", err)
		}
	}

	return &Handler{
		ch:    ch,
		log:   log.New(os.Stdout, "", log.LstdFlags),
		queue: queue,
		msgs:  msgs,
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

func (h *Handler) Withdraw(c *gin.Context) {
	var transaction tr.Transaction
	if err := c.BindJSON(&transaction); err != nil {
		h.log.Println("Withdraw:Error:", err)
		serverResponse(c, http.StatusInternalServerError, "", "", err.Error())
		return
	}

	corrId := shortid.MustGenerate()
	body, _ := json.Marshal(transaction)
	err := h.ch.Publish(
		"",
		"withdraw",
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       h.queue[0].Name,
			Body:          body,
		})
	if err != nil {
		h.log.Println("Withdraw:Error:", err.Error())
		serverResponse(c, http.StatusInternalServerError, "Failed to publish a message", "", err.Error())
		return
	}

	for d := range h.msgs[0] {
		h.log.Printf("%s: %s", d.CorrelationId, d.Body)
		if corrId == d.CorrelationId {
			err = json.Unmarshal(d.Body, &transaction)
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
	var transaction tr.Transaction
	if err := c.BindJSON(&transaction); err != nil {
		h.log.Println("Replenishment:Error:", err)
		serverResponse(c, http.StatusInternalServerError, "", "", err.Error())
		return
	}

	corrId := shortid.MustGenerate()
	body, _ := json.Marshal(transaction)
	err := h.ch.Publish(
		"",
		"replenishment",
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       h.queue[1].Name,
			Body:          body,
		})
	if err != nil {
		h.log.Println("Replenishment:Error:", err.Error())
		serverResponse(c, http.StatusInternalServerError, "Failed to publish a message", "", err.Error())
		return
	}

	for d := range h.msgs[1] {
		h.log.Printf("%s: %s", d.CorrelationId, d.Body)
		if corrId == d.CorrelationId {
			err = json.Unmarshal(d.Body, &transaction)
			if err != nil {
				h.log.Println("Replenishment:Error:", err.Error())
				serverResponse(c, http.StatusInternalServerError, "Failed to convert body to integer", "", err.Error())
				return
			}
			break
		}
	}

	serverResponse(c, http.StatusOK, "", transaction, "")
}

func (h *Handler) CardToCard(c *gin.Context) {
	var transaction tr.Transaction
	if err := c.BindJSON(&transaction); err != nil {
		h.log.Println("CardToCard:Error:", err)
		serverResponse(c, http.StatusInternalServerError, "", "", err.Error())
		return
	}

	corrId := shortid.MustGenerate()
	body, _ := json.Marshal(transaction)
	err := h.ch.Publish(
		"",
		"cardtocard",
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       h.queue[1].Name,
			Body:          body,
		})
	if err != nil {
		h.log.Println("CardToCard:Error:", err.Error())
		serverResponse(c, http.StatusInternalServerError, "Failed to publish a message", "", err.Error())
		return
	}

	for d := range h.msgs[1] {
		h.log.Printf("%s: %s", d.CorrelationId, d.Body)
		if corrId == d.CorrelationId {
			err = json.Unmarshal(d.Body, &transaction)
			if err != nil {
				h.log.Println("CardToCard:Error:", err.Error())
				serverResponse(c, http.StatusInternalServerError, "Failed to convert body to integer", "", err.Error())
				return
			}
			break
		}
	}

	serverResponse(c, http.StatusOK, "", transaction, "")
}
