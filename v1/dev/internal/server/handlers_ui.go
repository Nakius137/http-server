package server

import "net/http"

type homePageData struct {
	Username string
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	s.render(w, "home", homePageData{Username: "Guest"})
}
