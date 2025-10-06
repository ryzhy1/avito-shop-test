package main

import (
	"avito-shop/internal/app"
	"avito-shop/internal/config"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envDev   = "dev"
	envProd  = "prod"
	envLocal = "local"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(`                                                $$\                               $$\   
                                                $$ |                            $$$$ |  
 $$$$$$\        $$\   $$\       $$$$$$$$\       $$$$$$$\        $$\   $$\       \_$$ |  
$$  __$$\       $$ |  $$ |      \____$$  |      $$  __$$\       $$ |  $$ |        $$ |  
$$ |  \__|      $$ |  $$ |        $$$$ _/       $$ |  $$ |      $$ |  $$ |        $$ |  
$$ |            $$ |  $$ |       $$  _/         $$ |  $$ |      $$ |  $$ |        $$ |  
$$ |            \$$$$$$$ |      $$$$$$$$\       $$ |  $$ |      \$$$$$$$ |      $$$$$$\ 
\__|             \____$$ |      \________|      \__|  \__|       \____$$ |      \______|
                $$\   $$ |                                      $$\   $$ |              
                \$$$$$$  |                                      \$$$$$$  |              
                 \______/                                        \______/               `)

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	log.Info("Starting http", "env", cfg.Server.Env)

	application := app.New(
		log,
		cfg.Server.Address,
		cfg.Database.PostgresConn,
		cfg.JWT.Secret,
		cfg.JWT.AccessExpirationMinutes,
		cfg.JWT.RefreshExpirationDays,
	)

	go application.HTTPServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("Application stopped", slog.String("signal", sign.String()))

	err := application.HTTPServer.Stop(context.Background())
	if err != nil {
		return
	}
}
