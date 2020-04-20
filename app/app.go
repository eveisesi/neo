package app

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/killmail"
	"github.com/eveisesi/neo/services/market"
	"github.com/eveisesi/neo/services/token"
	"github.com/eveisesi/neo/services/universe"
	"golang.org/x/oauth2"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/eveisesi/neo/esi"
	"github.com/eveisesi/neo/mysql"
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	sqlDriver "github.com/go-sql-driver/mysql"
)

type App struct {
	Logger *logrus.Logger
	DB     *sqlx.DB
	Redis  *redis.Client
	Client *http.Client
	ESI    *esi.Client
	Config *neo.Config

	Alliance    alliance.Service
	Character   character.Service
	Corporation corporation.Service
	Killmail    killmail.Service
	Market      market.Service
	Token       token.Service
	Universe    universe.Service
}

func New() *App {

	cfg, err := loadEnv()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := makeLogger(cfg.LogLevel)
	if err != nil {
		if logger != nil {
			logger.WithError(err).Fatal("failed to configure logger")
		}
		log.Fatal(err)
	}

	db, err := makeDB(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("failed to make db connection")
	}

	logger.Info("pinging database server")

	err = db.Ping()
	if err != nil {
		logger.WithError(err).Fatal("failed to ping db server")
	}

	logger.Info("successfully pinged db server")

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	logger.Info("pinging redis server")

	pong, err := redisClient.Ping().Result()
	if err != nil {
		logger.WithError(err).Fatal("failed to ping redis server")
	}

	logger.WithField("pong", pong).Info("successfully pinged redis server")

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	esiClient := esi.New(client, cfg.ESIHost, cfg.ESIUAgent)

	txn := mysql.NewTransactioner(db)

	alliance := alliance.NewService(
		redisClient,
		esiClient,
		mysql.NewAllianceRepository(db),
	)
	character := character.NewService(
		redisClient,
		esiClient,
		mysql.NewCharacterRepository(db),
	)
	corporation := corporation.NewService(
		redisClient,
		esiClient,
		mysql.NewCorporationRepository(db),
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
		esiClient,
		logger,
		cfg,
		character,
		corporation,
		alliance,
		universe,
		market,
		txn,
		mysql.NewKillmailRepository(db),
	)

	return &App{
		Logger: logger,
		DB:     db,
		Redis:  redisClient,
		Client: client,
		ESI:    esiClient,
		Config: cfg,

		Alliance:    alliance,
		Character:   character,
		Corporation: corporation,
		Killmail:    killmail,
		Market:      market,
		Token:       token,
		Universe:    universe,
	}

}

func makeDB(cfg *neo.Config) (*sqlx.DB, error) {
	return mysql.Connect(&sqlDriver.Config{
		User:         cfg.DBUser,
		Passwd:       cfg.DBPass,
		Net:          "tcp",
		Addr:         cfg.DBHost,
		DBName:       cfg.DBName,
		Timeout:      time.Second * 2,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		ParseTime:    true,

		// Defaults
		Collation:            "utf8_general_ci",
		Loc:                  time.UTC,
		MaxAllowedPacket:     4 << 20, // 4 MiB
		AllowNativePasswords: true,
	})
}

func loadEnv() (*neo.Config, error) {
	config := neo.Config{}
	err := envconfig.Process("", &config)
	return &config, err
}

func makeLogger(logLevel string) (*logrus.Logger, error) {
	logger := logrus.New()

	logger.SetOutput(ioutil.Discard)

	logger.AddHook(&writerHook{
		Writer:    os.Stdout,
		LogLevels: logrus.AllLevels,
	})

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	logger.AddHook(&writerHook{
		Writer: &lumberjack.Logger{
			Filename: fmt.Sprintf("logs/%s/%s.log", hostname, time.Now().Format("2006-01-02T15:03:04")),
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
