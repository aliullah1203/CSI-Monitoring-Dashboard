package middlewares

import "backend/config"

type Middleware struct {
	cnf *config.Config
}

func NewMiddleware(cnf *config.Config) *Middleware {
	return &Middleware{cnf: cnf}
}
