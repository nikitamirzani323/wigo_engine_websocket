package router

import (
	"bufio"
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
	"github.com/valyala/fasthttp"
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
	app.Get("/sse", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			fmt.Println("WRITER")

			dbHost := os.Getenv("DB_REDIS_HOST") + ":" + os.Getenv("DB_REDIS_PORT")
			dbPass := os.Getenv("DB_REDIS_PASSWORD")
			dbName, _ := strconv.Atoi(os.Getenv("DB_REDIS_NAME"))

			rdb := redis.NewClient(&redis.Options{
				Addr:     dbHost,
				Password: dbPass,
				DB:       dbName,
			})

			resultredis := rdb.Subscribe("", "payload_nuke")
			defer resultredis.Close()
			for {
				msg, err := resultredis.ReceiveMessage()
				if err != nil {
					panic(err)
				}

				// fmt.Println("Received message from " + msg.Payload + " channel.")
				// data_pubsub := strings.Split(msg.Payload, ":")

				msg_sse := msg.Payload

				fmt.Fprintf(w, "data: %s\n\n", msg_sse)
				// fmt.Println(msg_sse)
				err_sse := w.Flush()
				if err_sse != nil {
					fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)

					break
				}
				// time.Sleep(1 * time.Second)

			}

		}))

		return nil
	})
	return app
}
func handleWebsocket(s *Server) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		fmt.Println("New incoming connection from client to orderbook feed:", ws.RemoteAddr())

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
func AllowUpgrade(ctx *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(ctx) {
		return ctx.Next()
	}
	return fiber.ErrUpgradeRequired
}
