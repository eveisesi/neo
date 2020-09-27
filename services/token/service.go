package token

import (
	"context"
	"net/http"
	"net/url"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// var ssoTimeLayout = "2006-01-02T15:04:05"

type Service interface {
	GetState(state string, scopes []string) string
	GetTokenForCode(ctx context.Context, state, code string) (*neo.Token, error)
	neo.TokenRepository
}

type service struct {
	client  *http.Client
	oauth   *oauth2.Config
	logger  *logrus.Logger
	redis   *redis.Client
	jwksURL string
	neo.TokenRepository
}

type (
	Request struct {
		Method  string
		Path    url.URL
		Headers map[string]string
		Body    []byte
	}

	Response struct {
		Method  string
		Path    string
		Code    int
		Headers map[string]string
		Data    interface{}
	}
)

func NewService(
	client *http.Client,
	oauth2 *oauth2.Config,
	logger *logrus.Logger,
	redis *redis.Client,
	jwksURL string,
	token neo.TokenRepository,
) Service {
	return &service{
		TokenRepository: token,
		client:          client,
		oauth:           oauth2,
		logger:          logger,
		redis:           redis,
		jwksURL:         jwksURL,
	}
}
