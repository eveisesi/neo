package app

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ddouglas/killboard/esi"
	"github.com/ddouglas/killboard/mysql"
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"

	"github.com/sirupsen/logrus"

	sqlDriver "github.com/go-sql-driver/mysql"
)

type App struct {
	Logger *logrus.Entry
	DB     *sqlx.DB
	Redis  *redis.Client
	Client *http.Client
	ESI    *esi.Client
	Config *config
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
}

func New() *App {

	cfg, err := loadEnv()
	if err != nil {
		log.Fatal(err)
	}

	logger := logrus.New()

	logger.SetOutput(os.Stdout)

	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.WithError(err).Fatal("failed to configure log level")
	}

	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

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

	host, err := os.Hostname()
	if err != nil {
		logger.WithError(err).Fatal("unable to determine hostname")
	}

	entry := logger.WithField("host", host)

	return &App{
		Logger: entry,
		DB:     db,
		Redis:  redisClient,
		Client: client,
		ESI:    esiClient,
		Config: &cfg,
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
