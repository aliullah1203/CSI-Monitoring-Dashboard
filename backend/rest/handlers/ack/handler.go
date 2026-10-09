package ack

import (
	"backend/repo"
	"backend/rest/middlewares"
)

type Handler struct {
	ackRepo     *repo.AckRepo
	middlewares *middlewares.Middleware
}

func NewHandler(ackRepo *repo.AckRepo, mw *middlewares.Middleware) *Handler {
	return &Handler{ackRepo: ackRepo, middlewares: mw}
}
