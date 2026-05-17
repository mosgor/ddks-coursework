package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Database Database `yaml:"database" env-prefix:"EVENTS_STORAGE_"`
	Server   Server   `yaml:"http_server" env-prefix:"EVENTS_SERVER_"`
	Redis    Redis    `yaml:"redis" env-prefix:"REDIS_"`
}

type Redis struct {
	Host     string `yaml:"host" env:"HOST" env-required:"false" env-description:"Redis host"`
	Port     string `yaml:"port" env:"PORT" env-required:"false" env-description:"Redis port"`
	Password string `yaml:"password" env:"PASSWORD" env-required:"false" env-description:"Redis password"`
	DB       int    `yaml:"db" env:"DB" env-default:"0" env-description:"Redis database number"`
}

type Database struct {
	Host     string `yaml:"host" env:"HOST" env-required:"true" env-description:"Database host"`
	Port     string `yaml:"port" env:"PORT" env-required:"true" env-description:"Database port"`
	User     string `yaml:"user" env:"USER" env-required:"true" env-description:"Database user"`
	Password string `yaml:"password" env:"PASSWORD" env-required:"true" env-description:"User's password"`
	Name     string `yaml:"name" env:"NAME" env-required:"true" env-description:"Table name"`

	MaxPoolSize  int           `yaml:"max_pool_size" env:"MAX_POOL_SIZE" env-default:"1" env-description:"Maximum amount of database connections"`
	ConnAttempts int           `yaml:"conn_attempts" env:"CONN_ATTEMPTS" env-default:"10" env-description:"How many times app will attempt to connect to the database"`
	Timeout      time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"1s" env-description:"Timeout before the next try to connect"`
}

type Server struct {
	JwtSecret string        `yaml:"jwt-secret" env:"JWT_SECRET" env-required:"true" env-description:"Secret for jwt tokens"`
	Timeout   time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"5s" env-description:"Connection timeout"`
	Port      string        `yaml:"port" env:"PORT" env-required:"true" env-description:"Server port"`
}

func NewConfig(configPath string) Config {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	// trying to read a config path from an environment variables if it's not specified before
	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	// using a default config path if it's not specified before
	if configPath == "" {
		logger.Warn("config path is not specified, using default \"config/config.yml\"")
		configPath = "config/config.yml"
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Error("config file does not exist: " + configPath)
		// panic("config file does not exist: " + *configPath)
	}

	cfg := Config{}
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		logger.Error("cannot read yaml: " + err.Error())
		// panic("cannot read yaml: " + err.Error())
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		logger.Error("cannot read env: " + err.Error())
		panic("cannot read env: " + err.Error())
	}

	return cfg
}
