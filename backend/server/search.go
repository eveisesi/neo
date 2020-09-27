package server

import (
	"errors"
	"net/http"
)

func (s *Server) handleSearchRequest(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()

	term := r.URL.Query().Get("term")
	if term == "" || len(term) <= 1 {
		s.WriteError(w, http.StatusBadRequest, errors.New("invalid term provided. Ensure term is atleast 2 characters long"))
		return
	}

	results, err := s.search.Fetch(ctx, term)
	if err != nil {
		s.WriteError(w, http.StatusBadRequest, errors.New("unable to complete the search request for that term"))
		return
	}

	s.WriteSuccess(w, http.StatusOK, results)
}
