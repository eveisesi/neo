package app

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

type App struct {
	esi *esi.Client
}

func New() {
	cfg, err := loadEnv()
	if err != nil {
		log.Fatal(err)
	}

	logger := logrus.New()

	logger.SetOutput(os.Stdout)

	level, err := logrus.ParseLevel(cfg.logLevel)
	if err != nil {
		logger.WithError(err).Fatal("failed to configure log level")
	}

	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})

}
