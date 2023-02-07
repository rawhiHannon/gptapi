package httpserver

import (
	"gptapi/internal/wsserver"
	"gptapi/pkg/utils"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type HttpServer struct {
	actions  map[string]utils.HttpHandler
	router   *mux.Router
	api      *mux.Router
	wsServer *wsserver.WsServer
}

func NewHttpServer() *HttpServer {
	networkInstance := &HttpServer{}
	networkInstance.init()
	networkInstance.initWebSocketsServer()

	return networkInstance
}

func (this *HttpServer) init() {
	this.actions = make(map[string]utils.HttpHandler)
	this.router = mux.NewRouter()
	this.api = this.router.PathPrefix("/api").Subrouter()
}

func (this *HttpServer) initWebSocketsServer() {
	this.wsServer = wsserver.NewWebsocketServer()

	this.api.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		this.wsServer.ServeWs(w, r)
	})
}

func (h *HttpServer) SetOnClientRegister(handler func(*wsserver.Client)) {
	h.wsServer.SetOnClientRegister(handler)
}

func (this *HttpServer) RegisterChannel(channel string) {
	if this.wsServer.HasRoom(channel) {
		return
	}
	this.wsServer.CreateRoom(channel)
}

func (h *HttpServer) Send(id string, data string) {
	h.wsServer.SendMessage(id, data)
}

func (this *HttpServer) Broadcast(channel string, data [][]string) {
	if this.wsServer.RoomHasListeners(channel) == false {
		return
	}
	go func() {
		var strData []string
		for _, arr := range data {
			strData = append(strData, strings.Join(arr, " "))
		}
		this.wsServer.BroadcastStream(channel, strings.Join(strData, "\n"))
	}()
}

func (this *HttpServer) BroadcastEvent(channel string, data string) {
	if this.wsServer.RoomHasListeners(channel) == false {
		return
	}
	this.wsServer.BroadcastEvent(channel, data)
}

func (this *HttpServer) Start(port string) {
	c := cors.New(cors.Options{
		AllowedHeaders:   []string{"X-Requested-With", "Content-Type", "Token", "Authorization", "X-Request-ID"},
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		// Debug:            true,
	})
	go this.wsServer.Run()
	http.ListenAndServe(":"+port, c.Handler(this.router))
}

// TODO: add another method that receive an array
func (this *HttpServer) RegisterMiddlewareAction(method string, route string, reqHandler utils.RequestHandler, middlewares []utils.Middleware) {
	handler := utils.GetHttpWrapper(reqHandler, middlewares)
	this.api.HandleFunc(route, handler).Methods(method)
}

func (this *HttpServer) RegisterAction(method string, route string, reqHandler utils.RequestHandler) {
	handler := utils.GetHttpWrapper(reqHandler, []utils.Middleware{})
	this.api.HandleFunc(route, handler).Methods(method)
}
