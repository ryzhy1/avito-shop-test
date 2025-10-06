package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	Env     string        `env:"ENV,required"` // local, dev, prod
	Address string        `env:"ADDRESS,required"`
	Timeout time.Duration `env:"TIMEOUT" envDefault:"5s"`
}

type DatabaseConfig struct {
	PostgresConn string `env:"POSTGRES_CONN,required"`
}

type JWTConfig struct {
	Secret                  string `env:"JWT_SECRET,required"`
	AccessExpirationMinutes int    `env:"ACCESS_EXPIRATION_MINUTES" envDefault:"15"`
	RefreshExpirationDays   int    `env:"REFRESH_EXPIRATION_DAYS" envDefault:"7"`
}

type RedisConfig struct {
	RedisConn string
}

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

const (
	local = ".env.local"
	dev   = ".env.dev"
	prod  = ".env.prod"
)

func MustLoad() *Config {
	if err := godotenv.Load(local); err != nil {
		panic(err)
	}

	timeoutStr := os.Getenv("TIMEOUT")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		panic("Invalid TIMEOUT format: " + err.Error())
	}

	accessExpStr := os.Getenv("ACCESS_EXPIRATION_MINUTES")
	accessExp, err := strconv.Atoi(accessExpStr)
	if err != nil {
		panic("Invalid ACCESS_EXPIRATION_MINUTES format: " + err.Error())
	}

	refreshExpStr := os.Getenv("REFRESH_EXPIRATION_DAYS")
	refreshExp, err := strconv.Atoi(refreshExpStr)
	if err != nil {
		panic("Invalid REFRESH_EXPIRATION_DAYS format: " + err.Error())
	}

	return &Config{
		Server: ServerConfig{
			Env:     os.Getenv("ENV"),
			Address: os.Getenv("ADDRESS"),
			Timeout: timeout,
		},
		Database: DatabaseConfig{
			PostgresConn: os.Getenv("POSTGRES_CONN"),
		},
		JWT: JWTConfig{
			Secret:                  os.Getenv("JWT_SECRET"),
			AccessExpirationMinutes: accessExp,
			RefreshExpirationDays:   refreshExp,
		},
	}
}
