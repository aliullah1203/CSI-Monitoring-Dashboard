package state

import (
	"backend/repo"
	"backend/rest/middlewares"
)

type Handler struct {
	stateRepo   *repo.StateRepo
	middlewares *middlewares.Middleware
}

func NewHandler(stateRepo *repo.StateRepo, mw *middlewares.Middleware) *Handler {
	return &Handler{stateRepo: stateRepo, middlewares: mw}
}
