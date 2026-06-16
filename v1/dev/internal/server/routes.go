package server

import (
	"net/http"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /users", s.handleGetUser)
	mux.HandleFunc("GET /{$}", s.handleHome)
	mux.HandleFunc("POST /users", s.handleAddUser)
	mux.HandleFunc("DELETE /users", s.handleDeleteUser)
	mux.HandleFunc("PUT /users", s.handleUpdateUser)

	return mux
}
