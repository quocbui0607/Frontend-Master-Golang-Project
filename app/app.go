package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"quocbui0607/femProject/api"
	"quocbui0607/femProject/migrations"
	"quocbui0607/femProject/stores"

	"github.com/joho/godotenv"
)

type Application struct {
	Logger         *log.Logger
	WorkoutHandler *api.WorkoutHandler
	DB             *sql.DB
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
	workoutHandler := api.NewWorkoutHandler(wkStore)
	app := &Application{
		Logger:         logger,
		WorkoutHandler: workoutHandler,
		DB:             pgDB,
	}

	return app, nil
}

func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "✅Health check")
}
