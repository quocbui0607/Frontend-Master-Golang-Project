package routes

import (
	"quocbui0607/femProject/app"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", app.HealthCheck)

	r.Get("/workouts/{id}", app.WorkoutHandler.HandleGetWorkoutByID)
	r.Post("/workouts", app.WorkoutHandler.HandleCreateWorkout)

	return r
}
