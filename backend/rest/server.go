package rest

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"backend/config"
	ackH "backend/rest/handlers/ack"
	eventsH "backend/rest/handlers/events"
	mqttstatusH "backend/rest/handlers/mqttstatus"
	stateH "backend/rest/handlers/state"
	"backend/rest/middlewares"
)

type Server struct {
	cnf           *config.Config
	eventsHandler *eventsH.Handler
	ackHandler    *ackH.Handler
	stateHandler  *stateH.Handler
}

func NewServer(
	cnf *config.Config,
	eventsHandler *eventsH.Handler,
	ackHandler *ackH.Handler,
	stateHandler *stateH.Handler,
) *Server {
	return &Server{
		cnf:           cnf,
		eventsHandler: eventsHandler,
		ackHandler:    ackHandler,
		stateHandler:  stateHandler,
	}
}

func (s *Server) Start() {
	manager := middlewares.NewManager()
	manager.Use(middlewares.Cors, middlewares.Preflight, middlewares.Logger)

	mux := http.NewServeMux()
	s.eventsHandler.RegisterRoutes(mux)
	s.ackHandler.RegisterRoutes(mux)
	s.stateHandler.RegisterRoutes(mux)
	mux.HandleFunc("GET /api/mqtt/status", mqttstatusH.GetStatus)

	wrappedMux := manager.WrapMux(mux)

	addr := ":" + strconv.Itoa(s.cnf.HttpPort)
	fmt.Println("🚀 Server running on", addr)
	if err := http.ListenAndServe(addr, wrappedMux); err != nil {
		fmt.Println("❌ Server error:", err)
		os.Exit(1)
	}
}
