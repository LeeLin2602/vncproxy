package server

import (
	"io"

	"time"
	"os"
	"net/http"
	"net/url"
	"github.com/amitbet/vncproxy/logger"

	"golang.org/x/net/websocket"
)

type WsServer struct {
	cfg *ServerConfig
}

type WsHandler func(io.ReadWriter, *ServerConfig, string)


func (wsServer *WsServer) Listen(urlStr string, handlerFunc WsHandler) {

	if urlStr == "" {
		urlStr = "/"
	}
	url, err := url.Parse(urlStr)
	if err != nil {
		logger.Errorf("error while parsing url: ", err)
	}

	connected := make(chan string, 5)
	singleConnection := 0

	http.Handle(url.Path, websocket.Handler(
		func(ws *websocket.Conn) {
			if singleConnection == 1 {
				return;
			}
			connected <- "connected"
			singleConnection = 1
			path := ws.Request().URL.Path
			var sessionId string
			if path != "" {
				sessionId = path[1:]
			}

			ws.PayloadType = websocket.BinaryFrame
			handlerFunc(ws, wsServer.cfg, sessionId)
			connected <- "exiting"
		}))

	go func(){
		err = http.ListenAndServe(url.Host, nil)
		if err != nil {
			panic("ListenAndServe: " + err.Error())
		}
	}()

	select {
	case <-connected:
		logger.Info("Connection Established.")
		// server.Shutdown(nil)
		for msg := range connected {
			logger.Infof("MSG: ", msg)
			if msg == "exiting" {
				os.Exit(0)
			}
		}
	case <-time.After(3 * time.Second):
		logger.Error("Timeout occurred. Stopping listener.")
		os.Exit(0)
	}
}
