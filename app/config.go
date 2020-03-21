package app

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	dbuser string `required:"true"`
	dbpass string `required:"true"`
	dbhost string `required:"true"`
	dbname string `required:"true"`

	logLevel string `required:"true"`
}

func loadEnv() (config config, err error) {
	_ = godotenv.Load(".env")

	err = envconfig.Process("", &config)

	return
}
