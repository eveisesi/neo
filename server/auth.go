package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eveisesi/neo/tools"
	"github.com/sirupsen/logrus"
)

func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {

	random := tools.RandomString(32)

	_, err := s.redis.Set(fmt.Sprintf("neo:state:%s", random), true, time.Minute*2).Result()
	if err != nil {
		_ = s.WriteError(w, http.StatusInternalServerError, errors.New("unable to handle request at this time"))
		return
	}

	scopes := make([]string, 0)
	if r.URL.Query().Get("scopes") != "" {
		scopes = strings.Split(r.URL.Query().Get("scopes"), ",")
	}

	url := s.token.GetState(random, scopes)

	_ = s.WriteSuccess(w, http.StatusOK, struct {
		URL string `json:"url"`
	}{
		URL: url,
	})

}

func (s *Server) handlePostCode(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	query := r.URL.Query()
	code := query.Get("code")
	state := query.Get("state")
	if code == "" || state == "" {
		_ = s.WriteError(w, http.StatusBadRequest, errors.New("code and state are required"))
		return
	}
	key := fmt.Sprintf("neo:state:%s", state)
	_, err := s.redis.Get(key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			err = errors.New("invalid state")
		}
		s.logger.WithError(err).WithFields(logrus.Fields{
			"code":  code,
			"state": state,
		}).Error("redis get error")
		_ = s.WriteError(w, http.StatusBadRequest, err)
		return
	}

	s.redis.Del(key)

	token, err := s.token.GetTokenForCode(ctx, state, code)
	if err != nil {
		msg := "failed to trade code for token"
		s.logger.WithError(err).Error(msg)
		_ = s.WriteError(w, http.StatusBadRequest, errors.New(msg))
		return
	}

	w.Header().Set("X-Neo-Token", token.AccessToken)
	w.WriteHeader(http.StatusNoContent)

}
