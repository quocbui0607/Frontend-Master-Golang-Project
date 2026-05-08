package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"quocbui0607/femProject/stores"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type WorkoutHandler struct {
	wkStore *stores.PostgresWorkoutStore
}

func NewWorkoutHandler(wkStore *stores.PostgresWorkoutStore) *WorkoutHandler {
	return &WorkoutHandler{wkStore: wkStore}
}

func (wh *WorkoutHandler) HandleGetWorkoutByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	workoutID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "this is workout id %d \n", workoutID)
}

func (wh *WorkoutHandler) HandleCreateWorkout(w http.ResponseWriter, r *http.Request) {
	var wk stores.Workout

	err := json.NewDecoder(r.Body).Decode(&wk)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to decode workout", http.StatusInternalServerError)
		return
	}

	createdWK, err := wh.wkStore.CreateWorkout(&wk)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to create workout", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdWK)
}
