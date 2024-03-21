package router

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	conns map[*websocket.Conn]bool
}

func Init() *fiber.App {
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(compress.New())
	// Custom config

	app.Use("ws/", AllowUpgrade)
	app.Use("ws/time", websocket.New(handleWebsocket(&Server{})))
	app.Use("ws/invoiceedit", websocket.New(handleWebsocketInvoiceEdit(&Server{})))

	return app
}
func handleWebsocket(s *Server) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		fmt.Println("New incoming connection from client to timer feed:", ws.RemoteAddr())

		msgType, mr, err := ws.ReadMessage()
		if err != nil {
			return
		}
		if msgType == websocket.TextMessage {
			// log.Println(string(mr))
		}
		keyredis := "payload_" + string(mr)
		dbHost := os.Getenv("DB_REDIS_HOST") + ":" + os.Getenv("DB_REDIS_PORT")
		dbPass := os.Getenv("DB_REDIS_PASSWORD")
		dbName, _ := strconv.Atoi(os.Getenv("DB_REDIS_NAME"))

		rdb := redis.NewClient(&redis.Options{
			Addr:     dbHost,
			Password: dbPass,
			DB:       dbName,
		})

		resultredis := rdb.Subscribe("", keyredis)
		defer resultredis.Close()

		for {
			msg, err := resultredis.ReceiveMessage()
			if err != nil {
				panic(err)
			}
			msg_sse := msg.Payload
			msg_replace := strings.Replace(msg_sse, `"`, "", 2)
			// fmt.Println(msg_replace)

			payload := fmt.Sprintf(msg_replace)
			ws.WriteJSON(payload)
			// time.Sleep(time.Second * 1)
			// err = ws.WriteMessage(msg_sse, []byte{msg_sse})
			// if err != nil {
			// 	log.Println("write:", err)
			// 	break
			// }

		}
	}
}
func handleWebsocketInvoiceEdit(s *Server) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		fmt.Println("New incoming connection from agen to invoice edit:", ws.RemoteAddr())

		msgType, mr, err := ws.ReadMessage()
		if err != nil {
			return
		}
		if msgType == websocket.TextMessage {
			// log.Println(string(mr))
		}
		keyredis := "payload_agen_" + string(mr)
		dbHost := os.Getenv("DB_REDIS_HOST") + ":" + os.Getenv("DB_REDIS_PORT")
		dbPass := os.Getenv("DB_REDIS_PASSWORD")
		dbName, _ := strconv.Atoi(os.Getenv("DB_REDIS_NAME"))

		rdb := redis.NewClient(&redis.Options{
			Addr:     dbHost,
			Password: dbPass,
			DB:       dbName,
		})

		resultredis := rdb.Subscribe("", keyredis)
		defer resultredis.Close()

		for {
			msg, err := resultredis.ReceiveMessage()
			if err != nil {
				panic(err)
			}
			msg_sse := msg.Payload
			msg_replace := strings.Replace(msg_sse, `"`, "", 2)
			// fmt.Println(msg_replace)

			payload := fmt.Sprintf(msg_replace)
			ws.WriteJSON(payload)
			// time.Sleep(time.Second * 1)
			// err = ws.WriteMessage(msg_sse, []byte{msg_sse})
			// if err != nil {
			// 	log.Println("write:", err)
			// 	break
			// }

		}
	}
}
func AllowUpgrade(ctx *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(ctx) {
		return ctx.Next()
	}
	return fiber.ErrUpgradeRequired
}
