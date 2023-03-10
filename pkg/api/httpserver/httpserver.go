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

func (h *HttpServer) init() {
	h.actions = make(map[string]utils.HttpHandler)
	h.router = mux.NewRouter()
	h.api = h.router.PathPrefix("/api").Subrouter()
}

func (h *HttpServer) initWebSocketsServer() {
	h.wsServer = wsserver.NewWebsocketServer()
	h.api.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		h.wsServer.ServeWs(w, r)
	})
}

func (h *HttpServer) SetOnClientRegister(handler func(*wsserver.Client)) {
	h.wsServer.SetOnClientRegister(handler)
}

func (h *HttpServer) RegisterChannel(channel string) {
	if h.wsServer.HasRoom(channel) {
		return
	}
	h.wsServer.CreateRoom(channel)
}

func (h *HttpServer) Send(id uint64, data string) {
	h.wsServer.SendMessage(id, data)
}

func (h *HttpServer) Broadcast(channel string, data [][]string) {
	if h.wsServer.RoomHasListeners(channel) == false {
		return
	}
	go func() {
		var strData []string
		for _, arr := range data {
			strData = append(strData, strings.Join(arr, " "))
		}
		h.wsServer.BroadcastStream(channel, strings.Join(strData, "\n"))
	}()
}

func (h *HttpServer) BroadcastEvent(channel string, data string) {
	if h.wsServer.RoomHasListeners(channel) == false {
		return
	}
	h.wsServer.BroadcastEvent(channel, data)
}

func (h *HttpServer) Start(port string) {
	c := cors.New(cors.Options{
		AllowedHeaders:   []string{"X-Requested-With", "Content-Type", "Token", "Authorization", "X-Request-ID"},
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		// Debug:            true,
	})
	go h.wsServer.Run()
	http.ListenAndServe(":"+port, c.Handler(h.router))
}

// TODO: add another method that receive an array
func (h *HttpServer) RegisterMiddlewareAction(method string, route string, reqHandler utils.RequestHandler, middlewares []utils.Middleware) {
	handler := utils.GetHttpWrapper(reqHandler, middlewares)
	h.api.HandleFunc(route, handler).Methods(method)
}

func (h *HttpServer) RegisterAction(method string, route string, reqHandler utils.RequestHandler) {
	handler := utils.GetHttpWrapper(reqHandler, []utils.Middleware{})
	h.api.HandleFunc(route, handler).Methods(method)
}

// func (h *HttpServer) AddNode(params map[string]string, queryString map[string][]string, bodyJson map[string]interface{}) (string, error) {
// 	log.Println("REST WORKS POST")
// 	return "", nil
// }
