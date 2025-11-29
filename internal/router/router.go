package router

import "net/http"

// postRouter captures the minimal interface we need for registering routes.
type postRouter interface {
	Post(string, http.HandlerFunc)
}

type Router struct{}

// RegisterRoutes wires up HTTP endpoints.
func (r *Router) RegisterRoutes(router postRouter) {
	router.Post("/api/auth/login", r.LoginHandler)
}
