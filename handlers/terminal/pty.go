package SATerminal

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gofiber/websocket/v2"
)

// TODO: support resize event of xterm.js

// type windowSize struct {
// 	Rows uint16 `json:"rows"`
// 	Cols uint16 `json:"cols"`
// 	X    uint16
// 	Y    uint16
// }

// func resizeHandler(c *websocket.Conn, reader *bytes.Reader, tty *os.File) {
// 	decoder := json.NewDecoder(reader)
// 	resizeMessage := windowSize{}
// 	err := decoder.Decode(&resizeMessage)
// 	if err != nil {
// 		c.WriteMessage(websocket.TextMessage, []byte("Error decoding resize message: "+err.Error()))
// 	}
// 	log.Println("resizeMessage", resizeMessage)
// 	_, _, errno := syscall.Syscall(
// 		syscall.SYS_IOCTL,
// 		tty.Fd(),
// 		syscall.TIOCSWINSZ,
// 		uintptr(unsafe.Pointer(&resizeMessage)),
// 	)
// 	if errno != 0 {
// 		log.Println("Unable to resize terminal", syscall.Errno(errno))
// 	}
// }

func CommandInteractive(c *websocket.Conn) {
	cmd := exec.Command("/bin/bash", "-l")
	cmd.Env = append(os.Environ(), "TERM=xterm")

	tty, err := pty.Start(cmd)
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		cmd.Process.Kill()
		cmd.Process.Wait()
		tty.Close()
		c.Close()
	}()

	quit := make(chan bool)

	go func() {
		select {
		case <-quit:
			return
		default:
			for {
				buf := make([]byte, 1024)
				read, err := tty.Read(buf)
				if err != nil {
					switch err.Error() {
					case "read /dev/ptmx: input/output error":
						if c != nil {
							c.WriteMessage(websocket.TextMessage, []byte("__AGENT_SIGNAL_CLOSE__"))
							c.Close()
						}
						return
					default:
						if c != nil {
							c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
							log.Print(err.Error())
						}
						return
					}
				}
				c.WriteMessage(websocket.BinaryMessage, buf[:read])
			}
		}
	}()

	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			log.Println(err.Error())
			if err.Error() == "websocket: close 1005 (no status)" {
				quit <- true
			}
			break
		}

		if mt != websocket.TextMessage {
			log.Printf("Error message format %d", mt)
		}

		if string(msg) == "__AGENT_SIGNAL_PING__" {
			log.Print("Received PING")
			c.WriteMessage(websocket.TextMessage, []byte("__AGENT_SIGNAL_PONG__"))
			continue
		}

		reader := bytes.NewReader([]byte(msg))
		copied, err := io.Copy(tty, reader)
		if err != nil {
			log.Printf("Error after copying %d bytes", copied)
		}
	}
}
