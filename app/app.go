package app

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/universe"

	"github.com/eveisesi/neo/esi"
	"github.com/eveisesi/neo/mysql"
	"github.com/eveisesi/neo/services/killmail"
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
	Config *config

	Alliance    alliance.Service
	Character   character.Service
	Corporation corporation.Service
	Killmail    killmail.Service
	Universe    universe.Service
}

type config struct {
	// db configuration
	DBUser string `required:"true"`
	DBPass string `required:"true"`
	DBHost string `required:"true"`
	DBName string `required:"true"`

	// logger configuration
	LogLevel string `required:"true"`

	// ESI configuration
	ESIHost   string `required:"true"`
	ESIUAgent string `required:"true"`

	// redis configuration
	RedisAddr string `required:"true"`

	// zkillboard params
	ZUAgent string `required:"true"`

	ServerPort uint `envconfig:"SERVER_PORT" required:"true"`
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
		Timeout: time.Second * 5,
	}

	esiClient := esi.New(client, cfg.ESIHost, cfg.ESIUAgent)

	alliance := alliance.NewService(mysql.NewAllianceRepository(db))
	character := character.NewService(mysql.NewCharacterRepository(db))
	corporation := corporation.NewService(mysql.NewCorporationRepository(db))
	killmail := killmail.NewService(mysql.NewKillmailRepository(db))
	universe := universe.NewService(mysql.NewUniverseRepository(db))

	return &App{
		Logger: logger,
		DB:     db,
		Redis:  redisClient,
		Client: client,
		ESI:    esiClient,
		Config: &cfg,

		Alliance:    alliance,
		Character:   character,
		Corporation: corporation,
		Killmail:    killmail,
		Universe:    universe,
	}

}

func makeDB(cfg config) (*sqlx.DB, error) {
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

func loadEnv() (config config, err error) {

	err = envconfig.Process("", &config)

	return
}

func makeLogger(logLevel string) (*logrus.Logger, error) {
	logger := logrus.New()

	logger.SetOutput(os.Stdout)

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
