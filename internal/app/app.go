package app

import (
	httpserver "avito-shop/internal/app/http-server"
	"avito-shop/internal/handlers"
	"avito-shop/internal/lib/jwt"
	"avito-shop/internal/middlewares"
	"avito-shop/internal/repository/postgres"
	"avito-shop/internal/repository/redis"
	"avito-shop/internal/routes"
	"avito-shop/internal/services"
	"context"
	"log/slog"
	"os"
	"time"
)

type App struct {
	HTTPServer *httpserver.Server
}

func New(log *slog.Logger, serverPort, storagePath, secret string, accessTTL, refreshTTL int) *App {
	storage, err := postgres.NewPostgres(context.Background(), storagePath)
	if err != nil {
		panic(err)
	}

	jwtGen := jwt.NewGenerator(secret, time.Minute*time.Duration(accessTTL), time.Hour*time.Duration(refreshTTL))

	redisDB, err := redis.InitRedis(os.Getenv("REDIS_STORAGE_PATH"), os.Getenv("redis_password"), os.Getenv("DB_NUMBER"), time.Duration(refreshTTL)*24)
	if err != nil {
		panic(err)
	}

	authService := services.NewAuthService(log, storage, redisDB, jwtGen)
	userService := services.NewUserService(log, storage)

	authHandler := handlers.NewAuthHandler(log, authService)
	userHandler := handlers.NewUserHandler(log, userService)

	authMiddleware := middlewares.NewAuthMiddleware(jwtGen)

	r := routes.InitRoutes(authHandler, userHandler, authMiddleware)

	server := httpserver.NewServer(log, serverPort, r)

	return &App{
		HTTPServer: server,
	}
}
