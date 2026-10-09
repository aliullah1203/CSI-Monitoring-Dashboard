package middlewares

import "net/http"

type Middlewares func(http.Handler) http.Handler

type Manager struct {
	globalMiddlwares []Middlewares
}

func NewManager() *Manager {
	return &Manager{globalMiddlwares: make([]Middlewares, 0)}
}

func (m *Manager) Use(middlewares ...Middlewares) {
	m.globalMiddlwares = append(m.globalMiddlwares, middlewares...)
}

func (m *Manager) WrapMux(next http.Handler) http.Handler {
	for _, mw := range m.globalMiddlwares {
		next = mw(next)
	}
	return next
}
