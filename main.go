package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"

	Metrics "server-agent/handlers/metrics"
	SAServers "server-agent/handlers/servers"
	Middlewares "server-agent/middlewares"
	Utils "server-agent/utils"

	SATerminal "server-agent/handlers/terminal"
)

type WSPongReply struct {
	MessageType string
}
type LoginStruct struct {
	User string `json:"user" xml:"user" form:"user"`
	Pass string `json:"pass" xml:"pass" form:"pass"`
}

func main() {
	cfg := &Utils.SAConfig{}
	if err := Utils.LoadConf("conf.yml", cfg); err != nil {
		log.Panicln(err)
	}

	var networkStr = fiber.NetworkTCP4
	if cfg.Feature.UseIpv6 {
		networkStr = fiber.NetworkTCP6
	}

	app := fiber.New(fiber.Config{
		Prefork: cfg.Feature.Prefork,
		Network: networkStr,
	})

	// Middlewares
	app.Use(cors.New())
	app.Use("/", func(c *fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/api") {
			// Except `/api/login`
			if c.Path() == "/api/login" {
				return c.Next()
			}
			exist_token := c.GetReqHeaders()["Server-Agent-Token"]
			if exist_token == "" {
				log.Printf("Request %s doesn't have token.", c.Path())
				if c.Method() == "GET" {
					if cfg.Feature.SinglePageUI {
						return c.Redirect("/login", 307)
					}
					return c.Redirect("/login.html", 307)
				}
				return c.SendStatus(403)
			}
			if exist_token != Utils.HmacSha256("admin:"+cfg.Basic.Pass, cfg.Basic.Salt) {
				log.Printf("Request %s is forbidden.", c.Path())
				if c.Method() == "GET" {
					if cfg.Feature.SinglePageUI {
						return c.Redirect("/login", 307)
					}
					return c.Redirect("/login.html", 307)
				}
				return c.SendStatus(403)
			}
			return c.Next()
		}
		if strings.HasPrefix(c.Path(), "/socket") {
			exist_token := c.Query("token")
			if exist_token == "" {
				log.Printf("Request %s doesn't have token.", c.Path())
				return c.SendStatus(403)
			}
			if exist_token != Utils.HmacSha256("admin:"+cfg.Basic.Pass, cfg.Basic.Salt) {
				log.Printf("Request %s is forbidden.", c.Path())
				return c.SendStatus(403)
			}
			return c.Next()
		}
		return c.Next()
	})
	app.Use("/socket", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// [Deprecated] HTTP fetch metrics with adaptor example
	// app.Get("/info", adaptor.HTTPHandler(Middlewares.ReplyInJson(func(w http.ResponseWriter, r *http.Request) {
	// 	response := Metrics.StaticMetricsJson()
	// 	fmt.Fprint(w, response)
	// })))

	app.Get("/api/server", func(c *fiber.Ctx) error {
		handler := c.Query("handler")

		switch handler {
		case "info":
			return c.SendString(Middlewares.Json2String(Metrics.StaticMetricsData()))
		case "state":
			return c.SendString(Middlewares.Json2String(Metrics.DynamicMetricsData()))
		default:
			return c.SendStatus(404)
		}
	})

	// Websocket fetch metrics
	app.Get("/socket/server", websocket.New(func(c *websocket.Conn) {
		// Access the websocket server: ws://localhost:3000/ws/123?v=1.0
		// c.Locals is added to the *websocket.Conn
		// log.Println(c.Locals("allowed"))  // true
		// log.Println(c.Params("id"))       // 123
		// log.Println(c.Query("v"))         // 1.0
		// log.Println(c.Cookies("session")) // ""

		var (
			mt  int
			msg []byte
			err error
		)

		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)

			switch string(msg) {
			case "info":
				c.WriteJSON(Metrics.StaticMetricsData())
			case "state":
				c.WriteJSON(Metrics.DynamicMetricsData())
			case "ping":
				resp := new(WSPongReply)
				resp.MessageType = "pong"
				c.WriteJSON(resp)
			default:
				c.Close()
			}

			if err = c.WriteMessage(mt, msg); err != nil {
				log.Println("write:", err)
				break
			}
		}
	}))

	app.Get("/socket/proxy", websocket.New(func(c *websocket.Conn) {
		// TODO: how to proxy websocket connection?
		proxyHost := c.Query("host")
		proxyPort := c.Query("port")
		proxyToken := c.Query("token")

		client := &http.Client{
			Timeout: time.Second * 120,
		}

		var (
			// mt  int
			msg []byte
			err error
		)
		for {
			if _, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)

			req, err := http.NewRequest("GET", "http://"+proxyHost+":"+proxyPort+"/api/server?handler="+string(msg), nil)
			if err != nil {
				log.Printf("Got error %s", err.Error())
				return
			}
			req.Header.Add("Server-Agent-Token", proxyToken)
			response, err := client.Do(req)
			if err != nil {
				log.Printf("Got error %s", err.Error())
				return
			}
			defer response.Body.Close()
			b, err := io.ReadAll(response.Body)
			if err != nil {
				log.Fatalln(err)
			}
			var jsonMap map[string]interface{}
			json.Unmarshal([]byte(string(b)), &jsonMap)
			c.WriteJSON(jsonMap)
		}
	}))

	// monitor's only routes
	if cfg.Basic.Mode == "monitor" {
		app.Post("/api/login", func(c *fiber.Ctx) error {
			p := new(LoginStruct)
			if err := c.BodyParser(p); err != nil {
				c.SendStatus(400)
				log.Print(err)
				return err
			}
			if p.User == "" && p.Pass == "" {
				c.SendStatus(400)
				return nil
			}
			if p.User != "admin" && p.Pass != cfg.Basic.Pass {
				c.SendStatus(403)
				return nil
			}
			c.SendString("{\"ok\":1,\"token\":\"" + Utils.HmacSha256(p.User+":"+p.Pass, cfg.Basic.Salt) + "\"}")
			return nil
		})
		app.Get("/api/list", func(c *fiber.Ctx) error {
			return c.SendString(Middlewares.Json2String(SAServers.LoadServerList()))
		})
		app.Get("/socket/terminal", websocket.New(func(c *websocket.Conn) {
			SATerminal.CommandInteractive(c)
		}))
		// Dashboard UI
		app.Static("/", "./pages")
		// Add fallback to `index.html` for SPA dashboard
		if cfg.Feature.SinglePageUI {
			log.Print("Dashboard is running in SPA mode.")
			app.Get("*", func(c *fiber.Ctx) error {
				return c.SendFile("./pages/index.html")
			})
		}
	}

	log.Fatal(app.Listen(":" + cfg.Basic.Port))
}
