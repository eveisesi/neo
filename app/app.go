package app

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/mdb"
	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/backup"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/esi"
	"github.com/eveisesi/neo/services/history"
	"github.com/eveisesi/neo/services/killmail"
	"github.com/eveisesi/neo/services/market"
	"github.com/eveisesi/neo/services/notifications"
	"github.com/eveisesi/neo/services/search"
	"github.com/eveisesi/neo/services/stats"
	"github.com/eveisesi/neo/services/token"
	"github.com/eveisesi/neo/services/top"
	"github.com/eveisesi/neo/services/tracker"
	"github.com/eveisesi/neo/services/universe"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/go-redis/redis/v7"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	"github.com/newrelic/go-agent/v3/newrelic"
)

type App struct {
	Label    string
	NewRelic *newrelic.Application
	Logger   *logrus.Logger
	MongoDB  *mongo.Database
	Redis    *redis.Client
	Client   *http.Client
	Config   *neo.Config
	Spaces   *session.Session
	ESI      esi.Service

	Alliance     alliance.Service
	Backup       backup.Service
	Character    character.Service
	Corporation  corporation.Service
	History      history.Service
	Killmail     killmail.Service
	Market       market.Service
	Search       search.Service
	Stats        stats.Service
	Notification notifications.Service
	Token        token.Service
	Top          top.Service
	Tracker      tracker.Service
	Universe     universe.Service
}

func New(command string, debug bool) *App {

	cfg, err := loadEnv()
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		cfg.LogLevel = "debug"
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	logger, err := makeLogger(hostname, command, cfg.LogLevel, cfg.Env)
	if err != nil {
		if logger != nil {
			logger.WithError(err).Fatal("failed to configure logger")
		}
		log.Fatal(err)
	}

	nr, err := makeNewRelicApp(cfg, logger, command)
	if err != nil {
		logger.WithError(err).Warn("failed to initialize newrelic application")
	}

	mongoDB, err := makeMongoDB(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("failed to make mongo db connection")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:               cfg.RedisAddr,
		MaxRetries:         3,
		IdleTimeout:        time.Second * 120,
		IdleCheckFrequency: time.Second * 10,
	})

	_, err = redisClient.Ping().Result()
	if err != nil {
		logger.WithError(err).Fatal("failed to ping redis server")
	}

	autocompleter := redisearch.NewAutocompleter(cfg.RedisAddr, "autocomplete")

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	client.Transport = newrelic.NewRoundTripper(client.Transport)

	esiClient := esi.New(redisClient, cfg.ESIHost, cfg.ESIUAgent)

	tracker := tracker.NewService(
		redisClient,
		logger,
	)

	alliance := alliance.NewService(
		redisClient,
		logger,
		nr,
		esiClient,
		tracker,
		mdb.NewAllianceRepository(mongoDB),
	)

	character := character.NewService(
		redisClient,
		logger,
		nr,
		esiClient,
		tracker,
		mdb.NewCharacterRepository(mongoDB),
	)

	corporation := corporation.NewService(
		redisClient,
		logger,
		nr,
		esiClient,
		tracker,
		mdb.NewCorporationRepository(mongoDB),
	)
	// TODO: Add support for search service back.
	// Need to replace data layer
	search := search.NewService(
		autocompleter,
		logger,
		mdb.NewCharacterRepository(mongoDB),
		mdb.NewCorporationRepository(mongoDB),
		mdb.NewAllianceRepository(mongoDB),
		mdb.NewUniverseRepository(mongoDB),
	)

	top := top.NewService(
		redisClient,
	)

	universe := universe.NewService(
		redisClient,
		esiClient,
		mdb.NewBlueprintRepository(mongoDB),
		mdb.NewUniverseRepository(mongoDB),
	)

	market := market.NewService(
		redisClient,
		esiClient,
		nr,
		logger,
		universe,
		mdb.NewMarketRepository(mongoDB),
		tracker,
	)

	// token := token.NewService(
	// 	client,
	// 	&oauth2.Config{
	// 		ClientID:     cfg.SSOClientID,
	// 		ClientSecret: cfg.SSOClientSecret,
	// 		RedirectURL:  cfg.SSOCallback,
	// 		Endpoint: oauth2.Endpoint{
	// 			AuthURL:  cfg.SSOAuthorizationURL,
	// 			TokenURL: cfg.SSOTokenURL,
	// 		},
	// 	},
	// 	logger,
	// 	redisClient,
	// 	cfg.SSOJWKSURL,
	// 	mysql.NewTokenRepository(mysqlDB),
	// )

	backup := backup.NewService(
		redisClient,
		logger,
	)

	killmail := killmail.NewService(
		client,
		redisClient,
		nr,
		esiClient,
		logger,
		cfg,
		backup,
		character,
		corporation,
		alliance,
		universe,
		market,
		tracker,
		mdb.NewKillmailRepository(mongoDB),
	)

	history := history.NewService(
		client,
		redisClient,
		logger,
		nr,
		cfg,
		mdb.NewKillmailRepository(mongoDB),
	)

	// stats := stats.NewService(redisClient, logger, nr, killmail, mysql.NewStatRepository(mysqlDB))

	notifications := notifications.NewService(
		client,
		redisClient,
		logger,
		nr,
		cfg,
		character,
		corporation,
		alliance,
		universe,
		killmail,
	)

	return &App{
		Label:    command,
		NewRelic: nr,
		Logger:   logger,
		MongoDB:  mongoDB,
		Redis:    redisClient,
		Client:   client,
		ESI:      esiClient,
		Config:   cfg,

		Alliance:     alliance,
		Backup:       backup,
		Character:    character,
		Corporation:  corporation,
		History:      history,
		Killmail:     killmail,
		Market:       market,
		Notification: notifications,
		Search:       search,
		// Stats:    stats,
		// Token:    token,
		Top:      top,
		Tracker:  tracker,
		Universe: universe,
	}

}

// makeNewRelicApp configures a instance of newrelic.Application for this application
// name is the command that this instance of the application is executing and is configured at runtime in func main
func makeNewRelicApp(cfg *neo.Config, logger *logrus.Logger, command string) (*newrelic.Application, error) {

	appName := cfg.NewRelicAppName
	if cfg.Env != "production" {
		appName = fmt.Sprintf("%s-%s", cfg.Env, appName)
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(cfg.NewRelicLicensenKey),
		newrelic.ConfigDistributedTracerEnabled(true),
		// newrelic.ConfigLogger(nrlogrus.Transform(logger)),
		// newrelic.ConfigDebugLogger(os.Stdout),
		func(config *newrelic.Config) {
			config.Labels = map[string]string{
				"command": command,
			}
		},
	)
	if err != nil {
		logger.WithError(err).Warn("failed to build newrelic application")
	}

	err = app.WaitForConnection(time.Second * 5)

	return app, err

}

func makeMongoDB(cfg *neo.Config) (*mongo.Database, error) {

	q := url.Values{}
	q.Set("authMechanism", cfg.Mongo.DBAuthMech)
	c := &url.URL{
		Scheme:   "mongodb",
		Host:     cfg.Mongo.DBHost,
		User:     url.UserPassword(cfg.Mongo.DBUser, cfg.Mongo.DBPass),
		Path:     fmt.Sprintf("/%s", cfg.Mongo.DBName),
		RawQuery: q.Encode(),
	}

	mc, err := mdb.Connect(context.TODO(), c)
	if err != nil {
		return nil, err
	}

	mdb := mc.Database(cfg.Mongo.DBName)

	return mdb, nil

}

func loadEnv() (*neo.Config, error) {
	config := neo.Config{}
	err := envconfig.Process("", &config)

	config.AllowedStatsEntities = []string{
		"character",
		"corporation",
		"alliance",

		"system",
		"constellation",
		"region",

		"ship",
	}

	return &config, err
}

func makeLogger(hostname, command, logLevel, env string) (*logrus.Logger, error) {
	logger := logrus.New()

	logger.SetOutput(ioutil.Discard)

	logger.AddHook(&writerHook{
		Writer:    os.Stdout,
		LogLevels: logrus.AllLevels,
	})

	logger.AddHook(&writerHook{
		Writer: &lumberjack.Logger{
			Filename: fmt.Sprintf("logs/error/%s-%s.log", hostname, command),
			MaxSize:  10,
			Compress: false,
		},
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
	})

	logger.AddHook(&writerHook{
		Writer: &lumberjack.Logger{
			Filename:   fmt.Sprintf("logs/info/%s-%s.log", hostname, command),
			MaxBackups: 3,
			MaxSize:    10,
			Compress:   false,
		},
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
		},
	})

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return logger, errors.Wrap(err, "failed to configure log level")
	}

	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return logger, err
}

type writerHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

func (w *writerHook) Fire(entry *logrus.Entry) error {

	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = w.Writer.Write([]byte(line))
	return err
}

func (w *writerHook) Levels() []logrus.Level {
	return w.LogLevels
}
