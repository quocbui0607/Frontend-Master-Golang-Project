package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"quocbui0607/femProject/api"
	"quocbui0607/femProject/middleware"
	"quocbui0607/femProject/migrations"
	"quocbui0607/femProject/stores"

	"github.com/joho/godotenv"
)

type Application struct {
	Logger            *log.Logger
	MiddlewareHandler middleware.UserMiddleware
	WorkoutHandler    *api.WorkoutHandler
	UserHandler       *api.UserHandler
	TokenHandler      *api.TokenHandler
	DB                *sql.DB
}

func NewApplication() (*Application, error) {
	envData, err := godotenv.Read(".env")
	if err != nil {
		return nil, fmt.Errorf("Error from loading env: %v", err)
	}

	ctx := context.WithValue(context.Background(), "env", envData)
	pgDB, err := stores.Open(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error from connecting db: %v", err)
	}

	err = stores.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	wkStore := stores.NewPostgresWorkoutStore(pgDB)
	userStore := stores.NewPostgresUserStore(pgDB)
	tokenStore := stores.NewPostgresTokenStore(pgDB)
	workoutHandler := api.NewWorkoutHandler(wkStore, logger)
	userHandler := api.NewUserHandler(userStore, logger)
	tokenHandler := api.NewTokenHandler(tokenStore, userStore, logger)
	middlewareHandler := middleware.UserMiddleware{UserStore: userStore}

	app := &Application{
		Logger:            logger,
		WorkoutHandler:    workoutHandler,
		UserHandler:       userHandler,
		TokenHandler:      tokenHandler,
		MiddlewareHandler: middlewareHandler,
		DB:                pgDB,
	}

	return app, nil
}

func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "✅Health check")
}
