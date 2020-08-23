package app

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/RediSearch/redisearch-go/redisearch"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/backup"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/esi"
	"github.com/eveisesi/neo/services/killmail"
	"github.com/eveisesi/neo/services/market"
	"github.com/eveisesi/neo/services/notifications"
	"github.com/eveisesi/neo/services/search"
	"github.com/eveisesi/neo/services/stats"
	"github.com/eveisesi/neo/services/token"
	"github.com/eveisesi/neo/services/top"
	"github.com/eveisesi/neo/services/tracker"
	"github.com/eveisesi/neo/services/universe"
	"golang.org/x/oauth2"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/eveisesi/neo/mysql"
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	"github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin"
	"github.com/newrelic/go-agent/v3/integrations/nrlogrus"
	"github.com/newrelic/go-agent/v3/newrelic"

	sqlDriver "github.com/go-sql-driver/mysql"
)

type App struct {
	Label    string
	NewRelic *newrelic.Application
	Logger   *logrus.Logger
	DB       *sqlx.DB
	Redis    *redis.Client
	Client   *http.Client
	Config   *neo.Config
	Spaces   *session.Session

	ESI          esi.Service
	Alliance     alliance.Service
	Backup       backup.Service
	Character    character.Service
	Corporation  corporation.Service
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

	logger, err := makeLogger(hostname, cfg.LogLevel, cfg.Env)
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

	db, err := makeDB(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("failed to make db connection")
	}

	err = db.Ping()
	if err != nil {
		logger.WithError(err).Fatal("failed to ping db server")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:               cfg.RedisAddr,
		MaxRetries:         3,
		IdleTimeout:        time.Second * 120,
		IdleCheckFrequency: time.Second * 10,
	})

	redisClient.AddHook(redisHook{
		cfg: cfg,
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

	esiClient := esi.New(client, redisClient, cfg.ESIHost, cfg.ESIUAgent)

	txn := mysql.NewTransactioner(db)

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
		mysql.NewAllianceRepository(db),
	)

	character := character.NewService(
		redisClient,
		logger,
		nr,
		esiClient,
		tracker,
		mysql.NewCharacterRepository(db),
	)

	corporation := corporation.NewService(
		redisClient,
		logger,
		nr,
		esiClient,
		tracker,
		mysql.NewCorporationRepository(db),
	)

	search := search.NewService(
		autocompleter,
		logger,
		mysql.NewSearchRepository(db),
	)

	top := top.NewService(
		redisClient,
	)

	universe := universe.NewService(
		redisClient,
		esiClient,
		mysql.NewBlueprintRepository(db),
		mysql.NewUniverseRepository(db),
	)

	market := market.NewService(
		redisClient,
		esiClient,
		logger,
		universe,
		txn,
		mysql.NewMarketRepository(db),
		tracker,
	)

	token := token.NewService(
		client,
		&oauth2.Config{
			ClientID:     cfg.SSOClientID,
			ClientSecret: cfg.SSOClientSecret,
			RedirectURL:  cfg.SSOCallback,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.SSOAuthorizationURL,
				TokenURL: cfg.SSOTokenURL,
			},
		},
		logger,
		redisClient,
		cfg.SSOJWKSURL,
		mysql.NewTokenRepository(db),
	)

	killmail := killmail.NewService(
		client,
		redisClient,
		nr,
		esiClient,
		logger,
		cfg,
		character,
		corporation,
		alliance,
		universe,
		market,
		tracker,
		txn,
		mysql.NewKillmailRepository(db),
		mysql.NewKillmailAttackerRepository(db),
		mysql.NewKillmailItemRepository(db),
		mysql.NewKillmailVictimRepository(db),
	)

	stats := stats.NewService(redisClient, logger, nr, killmail, mysql.NewStatRepository(db))

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

	spacesSession, e := session.NewSession(
		&aws.Config{
			Credentials: credentials.NewStaticCredentials(cfg.SpacesKey, cfg.SpacesSecret, ""),
			Endpoint:    aws.String(cfg.SpacesEndpoint),
			Region:      aws.String("us-east-1"),
		},
	)
	if e != nil {
		logger.WithError(err).Fatal("failed to create spaces session")
	}

	backup := backup.NewService(
		cfg.SpacesBucket,
		s3.New(spacesSession),
		redisClient,
		logger,
	)

	return &App{
		Label:    command,
		NewRelic: nr,
		Logger:   logger,
		DB:       db,
		Redis:    redisClient,
		Client:   client,
		ESI:      esiClient,
		Config:   cfg,
		Spaces:   spacesSession,

		Alliance:     alliance,
		Backup:       backup,
		Character:    character,
		Corporation:  corporation,
		Killmail:     killmail,
		Market:       market,
		Notification: notifications,
		Search:       search,
		Stats:        stats,
		Token:        token,
		Top:          top,
		Tracker:      tracker,
		Universe:     universe,
	}

}

// makeNewRelicApp configures a instance of newrelic.Application for this application
// name is the command that this instance of the application is executing and is configured at runtime in func main
func makeNewRelicApp(cfg *neo.Config, logger *logrus.Logger, command string) (*newrelic.Application, error) {

	env := ""
	if cfg.Env != "production" {
		env = cfg.Env
	}

	appName := fmt.Sprintf("%s-%s", env, cfg.NewRelicAppName)
	// appName := fmt.Sprintf("%s-%s-%s", env, cfg.NewRelicAppName, command)

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(cfg.NewRelicLicensenKey),
		newrelic.ConfigDistributedTracerEnabled(true),
		newrelic.ConfigLogger(nrlogrus.Transform(logger)),
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

func makeDB(cfg *neo.Config) (*sqlx.DB, error) {

	c := &sqlDriver.Config{
		User:         cfg.DBUser,
		Passwd:       cfg.DBPass,
		Net:          "tcp",
		Addr:         cfg.DBHost,
		DBName:       cfg.DBName,
		Timeout:      time.Second * 2,
		ReadTimeout:  time.Second * time.Duration(cfg.DBReadTimeout),
		WriteTimeout: time.Second * time.Duration(cfg.DBWriteTimeout),
		ParseTime:    true,

		// Defaults
		Collation:            "utf8_general_ci",
		Loc:                  time.UTC,
		MaxAllowedPacket:     4 << 20, // 4 MiB
		AllowNativePasswords: true,
	}

	db, err := mysql.Connect(c)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(64)
	db.SetMaxOpenConns(64)
	db.SetConnMaxLifetime(time.Minute * 5)

	return db, nil
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

func makeLogger(hostname, logLevel, env string) (*logrus.Logger, error) {
	logger := logrus.New()

	logger.SetOutput(ioutil.Discard)

	logger.AddHook(&writerHook{
		Writer:    os.Stdout,
		LogLevels: logrus.AllLevels,
	})

	logger.AddHook(&writerHook{
		Writer: &lumberjack.Logger{
			Filename: fmt.Sprintf("logs/%s/%s-error.log", hostname, time.Now().Format("2006-01-02T15:03:04")),
			MaxSize:  50,
			Compress: true,
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
			Filename:   fmt.Sprintf("logs/%s/%s-info.log", hostname, time.Now().Format("2006-01-02T15:03:04")),
			MaxBackups: 3,
			MaxSize:    10,
			Compress:   true,
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
	logger.SetFormatter(&nrlogrusplugin.ContextFormatter{})

	if env != "production" {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

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
