package events

import (
	"backend/repo"
	"backend/rest/middlewares"
)

type Handler struct {
	eventsRepo *repo.EventsRepo
	middlewares *middlewares.Middleware
}

func NewHandler(eventsRepo *repo.EventsRepo, mw *middlewares.Middleware) *Handler {
	return &Handler{eventsRepo: eventsRepo, middlewares: mw}
}
